package handler

import (
	"context"
	"errors"
	"fmt"
	"path"
	"strings"

	"go.uber.org/zap"

	"klubok/internal/application/service/image"
	"klubok/internal/application/zapctx"
	"klubok/internal/domain"
	"klubok/internal/restapi/oas"
)

const (
	imageMimePNG  = "image/png"
	imageMimeJPEG = "image/jpeg"

	// asciiFallbackName — имя файла в ASCII-fallback, когда исходное имя не ASCII.
	asciiFallbackName = "image"

	hexDigits = "0123456789ABCDEF"
)

type Image struct {
	diagramService    DiagramService
	imageService      ImageService
	permissionService PermissionService
	embedService      EmbedService

	log *zap.Logger
}

func NewImage(
	diagramService DiagramService,
	imageService ImageService,
	permissionService PermissionService,
	embedService EmbedService,
	log *zap.Logger,
) *Image {
	return &Image{
		diagramService:    diagramService,
		imageService:      imageService,
		permissionService: permissionService,
		embedService:      embedService,

		log: log.Named("image_handler"),
	}
}

// Upload image to diagram.
//
// POST /diagrams/{diagramId}/images.
func (h *Image) ImageUpload(
	ctx context.Context,
	req *oas.ImageUploadReq,
	params oas.ImageUploadParams,
) (oas.ImageUploadRes, error) {
	claims, err := domain.ClaimsFromContext(ctx)
	if err != nil {
		zapctx.Grasp(ctx, h.log).Error("Get claims from context",
			zap.Error(err),
		)
		return &oas.ImageUploadUnauthorized{}, nil
	}

	diagramID := domain.DiagramID(params.DiagramId)

	log := zapctx.Grasp(ctx, h.log).With(
		zap.String("diagram_id", diagramID.String()),
		zap.String("user_id", claims.UserID.String()),
	)

	// shortcut: существование диаграммы проверяется через Info, отдельного лёгкого
	// запроса в DiagramService нет; заменить, если загрузки станут частыми.
	if _, err := h.diagramService.Info(ctx, diagramID); err != nil {
		if errors.Is(err, domain.ErrDiagramNotFound) {
			log.Warn("Diagram not found")

			return &oas.ImageUploadNotFound{}, nil
		}

		log.Error("Get diagram info", zap.Error(err))

		return &oas.ImageUploadInternalServerError{}, nil
	}

	ok, err := h.permissionService.HasDiagramPermission(ctx, diagramID, claims.UserID, []domain.DiagramCollaboratorPermission{
		domain.DiagramCollaboratorPermissionOwner,
		domain.DiagramCollaboratorPermissionEdit,
	})
	if err != nil {
		log.Error("Check diagram collaborator permission",
			zap.Error(err),
		)
		return &oas.ImageUploadInternalServerError{}, nil
	}
	if !ok {
		log.Warn("User is not diagram owner or has not access edit to diagram")

		return &oas.ImageUploadForbidden{}, nil
	}

	if req.File.File == nil {
		log.Warn("Image file is empty")

		return &oas.ImageUploadBadRequest{}, nil
	}

	result, err := h.imageService.Upload(ctx, diagramID, req.File.Name, req.File.File)
	if err != nil {
		return imageUploadError(log, err)
	}

	return &oas.ImageUploadResponse{
		ID:           oas.Ulid(result.ID),
		OriginalName: result.OriginalName,
		Mime:         result.Mime,
		Width:        int32(result.Width),  //nolint:gosec // размеры ограничены MaxPixels
		Height:       int32(result.Height), //nolint:gosec // размеры ограничены MaxPixels
		Pv:           result.Pv,
	}, nil
}

func imageUploadError(log *zap.Logger, err error) (oas.ImageUploadRes, error) {
	switch {
	case errors.Is(err, image.ErrFileTooLarge):
		log.Warn("Image file too large", zap.Error(err))

		return &oas.ImageUploadRequestEntityTooLarge{}, nil
	case errors.Is(err, image.ErrInvalidImage):
		log.Warn("Image file is invalid", zap.Error(err))

		return &oas.ImageUploadBadRequest{}, nil
	default:
		log.Error("Upload image", zap.Error(err))

		return &oas.ImageUploadInternalServerError{}, nil
	}
}

// Get image variant of diagram.
//
// GET /diagrams/{diagramId}/images/{imageId}/{variant}.
func (h *Image) ImageGet(ctx context.Context, params oas.ImageGetParams) (oas.ImageGetRes, error) { //nolint:unparam // It's ok
	claims, err := domain.ClaimsFromContext(ctx)
	if err != nil {
		zapctx.Grasp(ctx, h.log).Error("Get claims from context",
			zap.Error(err),
		)
		return &oas.ImageGetUnauthorized{}, nil
	}

	diagramID := domain.DiagramID(params.DiagramId)

	log := zapctx.Grasp(ctx, h.log).With(
		zap.String("diagram_id", diagramID.String()),
		zap.String("image_id", string(params.ImageId)),
		zap.String("user_id", claims.UserID.String()),
	)

	variant, err := image.ParseVariant(string(params.Variant))
	if err != nil {
		log.Warn("Unknown image variant", zap.Error(err))

		return &oas.ImageGetBadRequest{}, nil
	}

	ok, err := h.permissionService.HasDiagramPermission(ctx, diagramID, claims.UserID, []domain.DiagramCollaboratorPermission{
		domain.DiagramCollaboratorPermissionOwner,
		domain.DiagramCollaboratorPermissionEdit,
		domain.DiagramCollaboratorPermissionView,
	})
	if err != nil {
		log.Error("Check diagram collaborator permission",
			zap.Error(err),
		)
		return &oas.ImageGetInternalServerError{}, nil
	}
	if !ok {
		log.Warn("User has not access to diagram")

		return &oas.ImageGetForbidden{}, nil
	}

	file, err := h.imageService.Open(ctx, diagramID, domain.ImageID(params.ImageId), variant)
	if err != nil {
		if errors.Is(err, image.ErrNotFound) {
			log.Warn("Image not found")

			return &oas.ImageGetNotFound{}, nil
		}

		log.Error("Open image variant", zap.Error(err))

		return &oas.ImageGetInternalServerError{}, nil
	}

	res, err := imageGetResponse(file, variant)
	if err != nil {
		// Тело ответа не отдаётся, поэтому reader до ogen не доходит — закрываем сами.
		_ = file.Reader.Close()

		log.Error("Build image response", zap.Error(err))

		return &oas.ImageGetInternalServerError{}, nil
	}

	return res, nil
}

// Get image variant by public embed token.
//
// GET /embed/{token}/images/{imageId}/{variant}.
func (h *Image) EmbedImageGet(ctx context.Context, params oas.EmbedImageGetParams) (oas.EmbedImageGetRes, error) { //nolint:unparam // It's ok
	log := zapctx.Grasp(ctx, h.log).With(
		zap.String("image_id", string(params.ImageId)),
		zap.String("variant", string(params.Variant)),
	)

	variant, err := image.ParseVariant(string(params.Variant))
	if err != nil {
		log.Warn("Unknown image variant", zap.Error(err))

		return &oas.EmbedImageGetBadRequest{}, nil
	}
	if variant != image.VariantFull {
		log.Warn("Image variant is not available in embed")

		return &oas.EmbedImageGetForbidden{}, nil
	}

	embed, err := h.embedService.GetByToken(ctx, params.Token)
	if err != nil {
		if errors.Is(err, domain.ErrEmbedNotFound) {
			log.Warn("Embed not found by token")

			return &oas.EmbedImageGetNotFound{}, nil
		}

		log.Error("Get embed by token", zap.Error(err))

		return &oas.EmbedImageGetInternalServerError{}, nil
	}

	file, err := h.imageService.Open(ctx, embed.DiagramID, domain.ImageID(params.ImageId), variant)
	if err != nil {
		if errors.Is(err, image.ErrNotFound) {
			log.Warn("Image not found")

			return &oas.EmbedImageGetNotFound{}, nil
		}

		log.Error("Open image variant", zap.Error(err))

		return &oas.EmbedImageGetInternalServerError{}, nil
	}

	res, err := embedImageGetResponse(file)
	if err != nil {
		// Тело ответа не отдаётся, поэтому reader до ogen не доходит — закрываем сами.
		_ = file.Reader.Close()

		log.Error("Build image response", zap.Error(err))

		return &oas.EmbedImageGetInternalServerError{}, nil
	}

	return res, nil
}

// imageGetResponse собирает ответ с байтами изображения; Content-Type выбирается
// по сохранённому MIME, размер отдаётся из хранилища.
func imageGetResponse(file image.File, variant image.Variant) (oas.ImageGetRes, error) {
	disposition := oas.NewOptString(contentDisposition(variant, file.OriginalName))
	length := oas.NewOptInt64(file.Size)

	switch file.Mime {
	case imageMimePNG:
		return &oas.ImageVariantsImagePNGHeaders{
			ContentDisposition: disposition,
			ContentLength:      length,
			Response:           oas.ImageVariantsImagePNG{Data: file.Reader},
		}, nil
	case imageMimeJPEG:
		return &oas.ImageVariantsImageJpegHeaders{
			ContentDisposition: disposition,
			ContentLength:      length,
			Response:           oas.ImageVariantsImageJpeg{Data: file.Reader},
		}, nil
	default:
		return nil, fmt.Errorf("unsupported image mime %q", file.Mime)
	}
}

func embedImageGetResponse(file image.File) (oas.EmbedImageGetRes, error) {
	res, err := imageGetResponse(file, image.VariantFull)
	if err != nil {
		return nil, err
	}

	// Типы ответа ImageVariants общие для обоих маршрутов.
	switch res := res.(type) {
	case *oas.ImageVariantsImagePNGHeaders:
		return res, nil
	case *oas.ImageVariantsImageJpegHeaders:
		return res, nil
	}

	return nil, fmt.Errorf("unsupported image response %T", res)
}

// contentDisposition собирает заголовок Content-Disposition: inline для просмотра,
// attachment для скачивания, с ASCII-fallback и filename* по RFC 5987.
func contentDisposition(variant image.Variant, name string) string {
	kind := "inline"
	if variant == image.VariantDownload {
		kind = "attachment"
	}

	return fmt.Sprintf(
		`%s; filename="%s"; filename*=UTF-8''%s`,
		kind,
		asciiFileName(name),
		rfc5987Encode(name),
	)
}

// asciiFileName оставляет осмысленное ASCII-имя: не-ASCII и кавычки заменяются
// на подчёркивания, имя без латинских букв и цифр получает fallback image.<ext>.
func asciiFileName(name string) string {
	ext := path.Ext(name)

	base := asciiPart(strings.TrimSuffix(name, ext))
	if !strings.ContainsFunc(base, isAlnum) {
		base = asciiFallbackName
	}

	return base + asciiPart(ext)
}

// asciiPart заменяет непечатаемые и не-ASCII символы одним подчёркиванием на серию.
func asciiPart(s string) string {
	var b strings.Builder

	b.Grow(len(s))

	underscore := false

	for _, r := range s {
		if r >= ' ' && r < 0x7f && r != '"' && r != '\\' {
			_, _ = b.WriteRune(r)

			underscore = false

			continue
		}

		if !underscore {
			_ = b.WriteByte('_')
			underscore = true
		}
	}

	return b.String()
}

func isAlnum(r rune) bool {
	return r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9'
}

// rfc5987Encode кодирует имя для filename*: всё вне attr-char (RFC 5987) в %XX.
func rfc5987Encode(name string) string {
	var b strings.Builder

	b.Grow(len(name))
	for i := range len(name) {
		c := name[i]
		if isAttrChar(c) {
			_ = b.WriteByte(c)
		} else {
			_ = b.WriteByte('%')
			_ = b.WriteByte(hexDigits[c>>4])
			_ = b.WriteByte(hexDigits[c&0x0f])
		}
	}

	if b.Len() == 0 {
		return asciiFallbackName
	}

	return b.String()
}

func isAttrChar(c byte) bool {
	switch {
	case c >= 'a' && c <= 'z', c >= 'A' && c <= 'Z', c >= '0' && c <= '9':
		return true
	case strings.IndexByte("!#$&+-.^_`|~", c) >= 0:
		return true
	default:
		return false
	}
}
