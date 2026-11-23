package handler

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	ht "github.com/ogen-go/ogen/http"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"klubok/internal/application/service/image"
	"klubok/internal/domain"
	"klubok/internal/restapi/handler/mocks"
	"klubok/internal/restapi/oas"
)

func claimsContext() context.Context {
	return domain.Claims{UserID: "user-1", SessionID: "sess-1"}.Context(context.Background())
}

func TestImage_ImageUpload(t *testing.T) {
	const diagramID = domain.DiagramID("diag-1")

	tests := []struct {
		name    string
		req     *oas.ImageUploadReq
		ctx     func() context.Context
		handler func(t *testing.T) *Image
		want    oas.ImageUploadRes
		wantErr bool
	}{
		{
			name: "Success",
			req: &oas.ImageUploadReq{
				File: oasFile("photo.png", "png bytes"),
			},
			ctx: claimsContext,
			handler: func(t *testing.T) *Image {
				t.Helper()

				return &Image{
					diagramService:    infoDiagram(t, diagramID, domain.DiagramInfo{ID: diagramID}, nil),
					permissionService: allowPermission(t, diagramID, true, nil),
					imageService: uploadImage(t, diagramID, "photo.png", "png bytes", image.UploadResult{
						ID:           "image-1",
						OriginalName: "photo.png",
						Mime:         "image/png",
						Width:        40,
						Height:       30,
						Pv:           "data:image/jpeg;base64,AAAA",
					}, nil),
					log: zap.NewNop(),
				}
			},
			want: &oas.ImageUploadResponse{
				ID:           "image-1",
				OriginalName: "photo.png",
				Mime:         "image/png",
				Width:        40,
				Height:       30,
				Pv:           "data:image/jpeg;base64,AAAA",
			},
		},
		{
			name: "Unauthorized",
			req: &oas.ImageUploadReq{
				File: oasFile("photo.png", "png bytes"),
			},
			ctx: context.Background,
			handler: func(t *testing.T) *Image {
				t.Helper()

				return &Image{log: zap.NewNop()}
			},
			want: &oas.ImageUploadUnauthorized{},
		},
		{
			name: "Diagram not found",
			req: &oas.ImageUploadReq{
				File: oasFile("photo.png", "png bytes"),
			},
			ctx: claimsContext,
			handler: func(t *testing.T) *Image {
				t.Helper()

				return &Image{
					diagramService: infoDiagram(t, diagramID, domain.DiagramInfo{}, domain.ErrDiagramNotFound),
					log:            zap.NewNop(),
				}
			},
			want: &oas.ImageUploadNotFound{},
		},
		{
			name: "Diagram info error",
			req: &oas.ImageUploadReq{
				File: oasFile("photo.png", "png bytes"),
			},
			ctx: claimsContext,
			handler: func(t *testing.T) *Image {
				t.Helper()

				return &Image{
					diagramService: infoDiagram(t, diagramID, domain.DiagramInfo{}, errServiceFailed),
					log:            zap.NewNop(),
				}
			},
			want: &oas.ImageUploadInternalServerError{},
		},
		{
			name: "Permission error",
			req: &oas.ImageUploadReq{
				File: oasFile("photo.png", "png bytes"),
			},
			ctx: claimsContext,
			handler: func(t *testing.T) *Image {
				t.Helper()

				return &Image{
					diagramService:    infoDiagram(t, diagramID, domain.DiagramInfo{ID: diagramID}, nil),
					permissionService: allowPermission(t, diagramID, false, errServiceFailed),
					log:               zap.NewNop(),
				}
			},
			want: &oas.ImageUploadInternalServerError{},
		},
		{
			name: "Forbidden without edit permission",
			req: &oas.ImageUploadReq{
				File: oasFile("photo.png", "png bytes"),
			},
			ctx: claimsContext,
			handler: func(t *testing.T) *Image {
				t.Helper()

				return &Image{
					diagramService:    infoDiagram(t, diagramID, domain.DiagramInfo{ID: diagramID}, nil),
					permissionService: allowPermission(t, diagramID, false, nil),
					log:               zap.NewNop(),
				}
			},
			want: &oas.ImageUploadForbidden{},
		},
		{
			name: "Empty file",
			req:  &oas.ImageUploadReq{},
			ctx:  claimsContext,
			handler: func(t *testing.T) *Image {
				t.Helper()

				return &Image{
					diagramService:    infoDiagram(t, diagramID, domain.DiagramInfo{ID: diagramID}, nil),
					permissionService: allowPermission(t, diagramID, true, nil),
					log:               zap.NewNop(),
				}
			},
			want: &oas.ImageUploadBadRequest{},
		},
		{
			name: "File too large",
			req: &oas.ImageUploadReq{
				File: oasFile("photo.png", "png bytes"),
			},
			ctx: claimsContext,
			handler: func(t *testing.T) *Image {
				t.Helper()

				return &Image{
					diagramService:    infoDiagram(t, diagramID, domain.DiagramInfo{ID: diagramID}, nil),
					permissionService: allowPermission(t, diagramID, true, nil),
					imageService:      uploadImage(t, diagramID, "photo.png", "png bytes", image.UploadResult{}, image.ErrFileTooLarge),
					log:               zap.NewNop(),
				}
			},
			want: &oas.ImageUploadRequestEntityTooLarge{},
		},
		{
			name: "Invalid image",
			req: &oas.ImageUploadReq{
				File: oasFile("photo.png", "png bytes"),
			},
			ctx: claimsContext,
			handler: func(t *testing.T) *Image {
				t.Helper()

				return &Image{
					diagramService:    infoDiagram(t, diagramID, domain.DiagramInfo{ID: diagramID}, nil),
					permissionService: allowPermission(t, diagramID, true, nil),
					imageService:      uploadImage(t, diagramID, "photo.png", "png bytes", image.UploadResult{}, image.ErrInvalidImage),
					log:               zap.NewNop(),
				}
			},
			want: &oas.ImageUploadBadRequest{},
		},
		{
			name: "Upload error",
			req: &oas.ImageUploadReq{
				File: oasFile("photo.png", "png bytes"),
			},
			ctx: claimsContext,
			handler: func(t *testing.T) *Image {
				t.Helper()

				return &Image{
					diagramService:    infoDiagram(t, diagramID, domain.DiagramInfo{ID: diagramID}, nil),
					permissionService: allowPermission(t, diagramID, true, nil),
					imageService:      uploadImage(t, diagramID, "photo.png", "png bytes", image.UploadResult{}, errServiceFailed),
					log:               zap.NewNop(),
				}
			},
			want: &oas.ImageUploadInternalServerError{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.handler(t).ImageUpload(tt.ctx(), tt.req, oas.ImageUploadParams{DiagramId: oas.Ulid(diagramID)})

			if tt.wantErr {
				require.Error(t, err)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestImage_ImageGet(t *testing.T) {
	const (
		diagramID = domain.DiagramID("diag-1")
		imageID   = domain.ImageID("image-1")
		content   = "png bytes"
	)

	tests := []struct {
		name    string
		params  oas.ImageGetParams
		ctx     func() context.Context
		handler func(t *testing.T) *Image
		want    oas.ImageGetRes
	}{
		{
			name:   "Full variant",
			params: oas.ImageGetParams{DiagramId: oas.Ulid(diagramID), ImageId: oas.Ulid(imageID), Variant: oas.ImageVariantEnumFull},
			ctx:    claimsContext,
			handler: func(t *testing.T) *Image {
				t.Helper()

				return &Image{
					permissionService: allowPermission(t, diagramID, true, nil),
					imageService:      openImage(t, diagramID, imageID, image.VariantFull, fileFixture("image/png", content), nil),
					log:               zap.NewNop(),
				}
			},
			want: &oas.ImageVariantsImagePNGHeaders{
				ContentDisposition: oas.NewOptString(`inline; filename="scheme.png"; filename*=UTF-8''scheme.png`),
				ContentLength:      oas.NewOptInt64(int64(len(content))),
				Response:           oas.ImageVariantsImagePNG{},
			},
		},
		{
			name:   "Download variant of non ASCII name",
			params: oas.ImageGetParams{DiagramId: oas.Ulid(diagramID), ImageId: oas.Ulid(imageID), Variant: oas.ImageVariantEnumDownload},
			ctx:    claimsContext,
			handler: func(t *testing.T) *Image {
				t.Helper()

				file := fileFixture("image/jpeg", content)
				file.OriginalName = "Схема.jpg"

				return &Image{
					permissionService: allowPermission(t, diagramID, true, nil),
					imageService:      openImage(t, diagramID, imageID, image.VariantDownload, file, nil),
					log:               zap.NewNop(),
				}
			},
			want: &oas.ImageVariantsImageJpegHeaders{
				ContentDisposition: oas.NewOptString(`attachment; filename="image.jpg"; filename*=UTF-8''%D0%A1%D1%85%D0%B5%D0%BC%D0%B0.jpg`),
				ContentLength:      oas.NewOptInt64(int64(len(content))),
				Response:           oas.ImageVariantsImageJpeg{},
			},
		},
		{
			name:   "Unauthorized",
			params: oas.ImageGetParams{DiagramId: oas.Ulid(diagramID), ImageId: oas.Ulid(imageID), Variant: oas.ImageVariantEnumFull},
			ctx:    context.Background,
			handler: func(t *testing.T) *Image {
				t.Helper()

				return &Image{log: zap.NewNop()}
			},
			want: &oas.ImageGetUnauthorized{},
		},
		{
			name:   "Unknown variant",
			params: oas.ImageGetParams{DiagramId: oas.Ulid(diagramID), ImageId: oas.Ulid(imageID), Variant: oas.ImageVariantEnum("preview")},
			ctx:    claimsContext,
			handler: func(t *testing.T) *Image {
				t.Helper()

				return &Image{log: zap.NewNop()}
			},
			want: &oas.ImageGetBadRequest{},
		},
		{
			name:   "Permission error",
			params: oas.ImageGetParams{DiagramId: oas.Ulid(diagramID), ImageId: oas.Ulid(imageID), Variant: oas.ImageVariantEnumFull},
			ctx:    claimsContext,
			handler: func(t *testing.T) *Image {
				t.Helper()

				return &Image{
					permissionService: allowPermission(t, diagramID, false, errServiceFailed),
					log:               zap.NewNop(),
				}
			},
			want: &oas.ImageGetInternalServerError{},
		},
		{
			name:   "Forbidden without view permission",
			params: oas.ImageGetParams{DiagramId: oas.Ulid(diagramID), ImageId: oas.Ulid(imageID), Variant: oas.ImageVariantEnumFull},
			ctx:    claimsContext,
			handler: func(t *testing.T) *Image {
				t.Helper()

				return &Image{
					permissionService: allowPermission(t, diagramID, false, nil),
					log:               zap.NewNop(),
				}
			},
			want: &oas.ImageGetForbidden{},
		},
		{
			name:   "Image not found",
			params: oas.ImageGetParams{DiagramId: oas.Ulid(diagramID), ImageId: oas.Ulid(imageID), Variant: oas.ImageVariantEnumFull},
			ctx:    claimsContext,
			handler: func(t *testing.T) *Image {
				t.Helper()

				return &Image{
					permissionService: allowPermission(t, diagramID, true, nil),
					imageService:      openImage(t, diagramID, imageID, image.VariantFull, image.File{}, image.ErrNotFound),
					log:               zap.NewNop(),
				}
			},
			want: &oas.ImageGetNotFound{},
		},
		{
			name:   "Storage error",
			params: oas.ImageGetParams{DiagramId: oas.Ulid(diagramID), ImageId: oas.Ulid(imageID), Variant: oas.ImageVariantEnumFull},
			ctx:    claimsContext,
			handler: func(t *testing.T) *Image {
				t.Helper()

				return &Image{
					permissionService: allowPermission(t, diagramID, true, nil),
					imageService:      openImage(t, diagramID, imageID, image.VariantFull, image.File{}, errServiceFailed),
					log:               zap.NewNop(),
				}
			},
			want: &oas.ImageGetInternalServerError{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.handler(t).ImageGet(tt.ctx(), tt.params)
			require.NoError(t, err)
			assertImageVariant(t, got, tt.want, content)
		})
	}
}

func TestImage_EmbedImageGet(t *testing.T) {
	const (
		diagramID = domain.DiagramID("diag-1")
		imageID   = domain.ImageID("image-1")
		token     = "token"
		content   = "png bytes"
	)

	tests := []struct {
		name    string
		params  oas.EmbedImageGetParams
		handler func(t *testing.T) *Image
		want    oas.EmbedImageGetRes
	}{
		{
			name:   "Full variant",
			params: oas.EmbedImageGetParams{Token: token, ImageId: oas.Ulid(imageID), Variant: oas.ImageVariantEnumFull},
			handler: func(t *testing.T) *Image {
				t.Helper()

				return &Image{
					embedService: embedByToken(t, token, diagramID, nil),
					imageService: openImage(t, diagramID, imageID, image.VariantFull, fileFixture("image/png", content), nil),
					log:          zap.NewNop(),
				}
			},
			want: &oas.ImageVariantsImagePNGHeaders{
				ContentDisposition: oas.NewOptString(`inline; filename="scheme.png"; filename*=UTF-8''scheme.png`),
				ContentLength:      oas.NewOptInt64(int64(len(content))),
				Response:           oas.ImageVariantsImagePNG{},
			},
		},
		{
			name:   "Download is forbidden in embed",
			params: oas.EmbedImageGetParams{Token: token, ImageId: oas.Ulid(imageID), Variant: oas.ImageVariantEnumDownload},
			handler: func(t *testing.T) *Image {
				t.Helper()

				return &Image{log: zap.NewNop()}
			},
			want: &oas.EmbedImageGetForbidden{},
		},
		{
			name:   "Preview is rejected in embed",
			params: oas.EmbedImageGetParams{Token: token, ImageId: oas.Ulid(imageID), Variant: oas.ImageVariantEnum("preview")},
			handler: func(t *testing.T) *Image {
				t.Helper()

				return &Image{log: zap.NewNop()}
			},
			want: &oas.EmbedImageGetBadRequest{},
		},
		{
			name:   "Revoked token",
			params: oas.EmbedImageGetParams{Token: token, ImageId: oas.Ulid(imageID), Variant: oas.ImageVariantEnumFull},
			handler: func(t *testing.T) *Image {
				t.Helper()

				return &Image{
					embedService: embedByToken(t, token, diagramID, domain.ErrEmbedNotFound),
					log:          zap.NewNop(),
				}
			},
			want: &oas.EmbedImageGetNotFound{},
		},
		{
			name:   "Embed service error",
			params: oas.EmbedImageGetParams{Token: token, ImageId: oas.Ulid(imageID), Variant: oas.ImageVariantEnumFull},
			handler: func(t *testing.T) *Image {
				t.Helper()

				return &Image{
					embedService: embedByToken(t, token, diagramID, errServiceFailed),
					log:          zap.NewNop(),
				}
			},
			want: &oas.EmbedImageGetInternalServerError{},
		},
		{
			name:   "Image of another diagram",
			params: oas.EmbedImageGetParams{Token: token, ImageId: oas.Ulid(imageID), Variant: oas.ImageVariantEnumFull},
			handler: func(t *testing.T) *Image {
				t.Helper()

				return &Image{
					embedService: embedByToken(t, token, diagramID, nil),
					imageService: openImage(t, diagramID, imageID, image.VariantFull, image.File{}, image.ErrNotFound),
					log:          zap.NewNop(),
				}
			},
			want: &oas.EmbedImageGetNotFound{},
		},
		{
			name:   "Storage error",
			params: oas.EmbedImageGetParams{Token: token, ImageId: oas.Ulid(imageID), Variant: oas.ImageVariantEnumFull},
			handler: func(t *testing.T) *Image {
				t.Helper()

				return &Image{
					embedService: embedByToken(t, token, diagramID, nil),
					imageService: openImage(t, diagramID, imageID, image.VariantFull, image.File{}, errServiceFailed),
					log:          zap.NewNop(),
				}
			},
			want: &oas.EmbedImageGetInternalServerError{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.handler(t).EmbedImageGet(context.Background(), tt.params)
			require.NoError(t, err)
			assertImageVariant(t, got, tt.want, content)
		})
	}
}

func TestContentDisposition(t *testing.T) {
	tests := []struct {
		name    string
		variant image.Variant
		file    string
		want    string
	}{
		{
			name:    "Inline ASCII name",
			variant: image.VariantFull,
			file:    "scheme.png",
			want:    `inline; filename="scheme.png"; filename*=UTF-8''scheme.png`,
		},
		{
			name:    "Attachment ASCII name",
			variant: image.VariantDownload,
			file:    "scheme-1_v2.png",
			want:    `attachment; filename="scheme-1_v2.png"; filename*=UTF-8''scheme-1_v2.png`,
		},
		{
			name:    "Non ASCII name keeps ASCII fallback",
			variant: image.VariantDownload,
			file:    "схема.png",
			want:    `attachment; filename="image.png"; filename*=UTF-8''%D1%81%D1%85%D0%B5%D0%BC%D0%B0.png`,
		},
		{
			name:    "Quotes and backslash are neutralized",
			variant: image.VariantFull,
			file:    `a"b\c.png`,
			want:    `inline; filename="a_b_c.png"; filename*=UTF-8''a%22b%5Cc.png`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, contentDisposition(tt.variant, tt.file))
		})
	}
}

var errServiceFailed = errors.New("service failed")

func oasFile(name, content string) ht.MultipartFile {
	return ht.MultipartFile{Name: name, File: strings.NewReader(content)}
}

func fileFixture(mime, content string) image.File {
	return image.File{
		Reader:       io.NopCloser(strings.NewReader(content)),
		Size:         int64(len(content)),
		Mime:         mime,
		OriginalName: "scheme.png",
	}
}

func infoDiagram(
	t *testing.T,
	diagramID domain.DiagramID,
	info domain.DiagramInfo,
	err error,
) *mocks.DiagramServiceMock {
	t.Helper()

	svc := mocks.NewDiagramServiceMock(t)
	svc.EXPECT().Info(mock.Anything, diagramID).Return(info, err)

	return svc
}

func allowPermission(
	t *testing.T,
	diagramID domain.DiagramID,
	allowed bool,
	err error,
) *mocks.PermissionServiceMock {
	t.Helper()

	svc := mocks.NewPermissionServiceMock(t)
	svc.EXPECT().
		HasDiagramPermission(mock.Anything, diagramID, domain.UserID("user-1"), mock.Anything).
		Return(allowed, err)

	return svc
}

func uploadImage(
	t *testing.T,
	diagramID domain.DiagramID,
	name string,
	content string,
	result image.UploadResult,
	err error,
) *mocks.ImageServiceMock {
	t.Helper()

	svc := mocks.NewImageServiceMock(t)
	svc.EXPECT().
		Upload(mock.Anything, diagramID, name, mock.MatchedBy(func(r io.Reader) bool {
			data, readErr := io.ReadAll(r)

			return readErr == nil && string(data) == content
		})).
		Return(result, err)

	return svc
}

func openImage(
	t *testing.T,
	diagramID domain.DiagramID,
	imageID domain.ImageID,
	variant image.Variant,
	file image.File,
	err error,
) *mocks.ImageServiceMock {
	t.Helper()

	svc := mocks.NewImageServiceMock(t)
	svc.EXPECT().Open(mock.Anything, diagramID, imageID, variant).Return(file, err)

	return svc
}

func embedByToken(
	t *testing.T,
	token string,
	diagramID domain.DiagramID,
	err error,
) *mocks.EmbedServiceMock {
	t.Helper()

	svc := mocks.NewEmbedServiceMock(t)
	svc.EXPECT().
		GetByToken(mock.Anything, token).
		Return(domain.DiagramEmbedInfo{DiagramID: diagramID}, err)

	return svc
}

// assertImageVariant сравнивает ответ с байтами изображения, не сравнивая reader напрямую.
func assertImageVariant(t *testing.T, got, want any, body string) {
	t.Helper()

	switch wantPNG := want.(type) {
	case *oas.ImageVariantsImagePNGHeaders:
		gotPNG, ok := got.(*oas.ImageVariantsImagePNGHeaders)
		require.True(t, ok)

		assert.Equal(t, wantPNG.ContentDisposition, gotPNG.ContentDisposition)
		assert.Equal(t, wantPNG.ContentLength, gotPNG.ContentLength)
		assert.Equal(t, body, readBody(t, gotPNG.Response.Data))

	case *oas.ImageVariantsImageJpegHeaders:
		gotJPEG, ok := got.(*oas.ImageVariantsImageJpegHeaders)
		require.True(t, ok)

		assert.Equal(t, wantPNG.ContentDisposition, gotJPEG.ContentDisposition)
		assert.Equal(t, wantPNG.ContentLength, gotJPEG.ContentLength)
		assert.Equal(t, body, readBody(t, gotJPEG.Response.Data))

	default:
		assert.Equal(t, want, got)
	}
}

func readBody(t *testing.T, r io.Reader) string {
	t.Helper()

	data, err := io.ReadAll(r)
	require.NoError(t, err)

	return string(data)
}
