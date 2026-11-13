/* eslint-disable */
/* tslint:disable */
// @ts-nocheck
/*
 * ---------------------------------------------------------------
 * ## THIS FILE WAS GENERATED VIA SWAGGER-TYPESCRIPT-API        ##
 * ##                                                           ##
 * ## AUTHOR: acacode                                           ##
 * ## SOURCE: https://github.com/acacode/swagger-typescript-api ##
 * ---------------------------------------------------------------
 */

export type DiagramCollaboratorPermissionEnum = "OWNER" | "VIEW" | "EDIT";

/**
 * - S: SQUARE
 * - E: ELLIPSE
 * - D: DIAMOND
 * - B: BOX
 * - T: TEXT
 */
export type DiagramShapeTypeEnum = "S" | "E" | "D" | "B" | "T";

/**
 * ID
 * @pattern ^[0-9A-HJKMNP-TV-Z]{26}$
 * @example "01KM5CJP5ZHAGX8MJ0CN6988AF"
 */
export type Ulid = string;

/**
 * User email address
 * @minLength 3
 * @maxLength 255
 * @pattern ^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$
 * @example "user@domain.tld"
 */
export type Email = string;

/**
 * Username
 * @minLength 1
 * @maxLength 50
 * @pattern ^[a-z0-9]+$
 * @example "dimnik"
 */
export type Username = string;

/**
 * Diagram title
 * @minLength 1
 * @maxLength 255
 */
export type DiagramTitle = string;

/**
 * Datetime
 * @format date-time
 * @example "2020-01-01T00:00:00.000Z"
 */
export type Datetime = string;

export interface GeneralError {
  message?: string | null;
}

export interface EmbedInfoResponse {
  embed: EmbedInfo;
}

export interface EmbedListResponse {
  embeds: Array<EmbedInfo>;
}

export interface EmbedDataResponse {
  diagramInfo: DiagramInfo;
}

export interface IndexResponse {
  name: string;
}

export interface FeaturesResponse {
  captcha: FeaturesCaptcha;
  authEmail: FeaturesAuthEmail;
  authSso: FeaturesAuthSso;
}

export interface DiagramResponse {
  diagram: Diagram;
}

export interface DiagramListResponse {
  diagrams: Array<Diagram>;
}

export interface DiagramCreateRequest {
  /** Diagram title */
  title: DiagramTitle;
}

export interface DiagramImportRequest {
  /** Diagram title */
  title: DiagramTitle;
  /** @minItems 1 */
  shapes: Array<DiagramShape>;
  connections: Array<DiagramConnection>;
}

export interface DiagramSetTitleRequest {
  /** Diagram title */
  title: DiagramTitle;
}

export interface DiagramInfoResponse {
  diagramInfo: DiagramInfo;
}

export interface DiagramStatResponse {
  /** @format int32 */
  websocketConnections: number;
  cursorPositions: Array<CursorPosition>;
}

export interface DiagramOnlineUsersResponse {
  onlineUsers: Array<OnlineUser>;
}

export interface CollaboratorListResponse {
  collaborators: Array<DiagramCollaborator>;
}

export interface CollaboratorUpsertRequest {
  /**
   * User name or email
   * @minLength 1
   * @maxLength 255
   */
  userNameOrEmail: string;
  permission: DiagramCollaboratorPermissionEnum;
}

export interface CollaboratorUpsertResponse {
  collaborator: DiagramCollaborator;
}

export interface ShapeUpsertRequest {
  shape: DiagramShape;
}

export interface ShapeBatchUpsertRequest {
  /**
   * @maxItems 100
   * @minItems 1
   */
  shapes: Array<DiagramShape>;
}

export interface ConnectionUpsertRequest {
  connection: DiagramConnection;
}

export interface ConnectionBatchUpsertRequest {
  /**
   * @maxItems 100
   * @minItems 1
   */
  connections: Array<DiagramConnection>;
}

export interface CommentListResponse {
  comments: Array<DiagramComment>;
}

export interface CommentCreateRequest {
  /**
   * Comment text
   * @minLength 1
   */
  text: string;
}

export interface CommentCreateResponse {
  comment: DiagramComment;
}

export interface UserRegisterRequest {
  /** User email address */
  email: Email;
  /** Username */
  username: Username;
  /** Captcha token */
  captchaToken: string;
}

export interface UserRegisterResponse {
  /** ID */
  id: Ulid;
  /** User email address */
  email: Email;
  /** Username */
  username: Username;
}

export interface UserActivateRequest {
  /** User ID */
  userId: Ulid;
  /**
   * User activation code
   * @minLength 1
   */
  code: string;
}

export interface UserLoginRequest {
  /** User email address */
  email: Email;
  /** Captcha token */
  captchaToken: string;
}

export interface UserLoginVerifyRequest {
  /** User email address */
  email: Email;
  /** User verification code */
  code: string;
}

export interface UserMeResponse {
  /** ID */
  id: Ulid;
  /** User email address */
  email: Email;
  /** Username */
  username: Username;
  /** Whether the account is an automatically created fake account */
  isFake: boolean;
}

export interface FeaturesCaptcha {
  /** Captcha is enabled. */
  enabled: boolean;
  /** Public captcha site key. */
  siteKey: string;
}

export interface FeaturesAuthEmail {
  /** Email login is enabled. */
  enabled: boolean;
}

export interface FeaturesAuthSso {
  /** SSO login is enabled. */
  enabled: boolean;
}

export interface Diagram {
  /** ID */
  id: Ulid;
  /** Diagram title */
  title: DiagramTitle;
  permission: DiagramCollaboratorPermissionEnum;
  hasCollaborators: boolean;
  /** Datetime */
  createdAt: Datetime;
  /** Datetime */
  updatedAt: Datetime;
}

export interface DiagramInfo {
  /** ID */
  id: Ulid;
  /** Diagram title */
  title: DiagramTitle;
  permission: DiagramCollaboratorPermissionEnum;
  shapes: Array<DiagramShape>;
  connections: Array<DiagramConnection>;
  shapeCommentStats: Array<DiagramShapeCommentStats>;
}

export interface DiagramShapeCommentStats {
  /** ID */
  shapeId: Ulid;
  /** @format int32 */
  commentCount: number;
}

export interface DiagramShape {
  /** ID. */
  id: Ulid;
  tp: DiagramShapeTypeEnum;
  /**
   * Left.
   * @format double
   */
  l: number;
  /**
   * Top.
   * @format double
   */
  t: number;
  /**
   * Width.
   * @format double
   */
  w: number;
  /**
   * Height.
   * @format double
   */
  h: number;
  /**
   * Scale X.
   * @format double
   */
  sx: number;
  /**
   * Scale Y.
   * @format double
   */
  sy: number;
  /**
   * Fill color. Accepts HEX color code or "transparent".
   * @pattern ^(#[0-9A-Fa-f]{6}|transparent)$
   * @example "#000000"
   */
  fc: string;
  /**
   * Stroke color. Accepts HEX color code or "transparent".
   * @pattern ^(#[0-9A-Fa-f]{6}|transparent)$
   * @example "#000000"
   */
  sc: string;
  /**
   * Text.
   * @minLength 1
   */
  txt: string;
  /** Order key for layering */
  o: string;
}

export interface DiagramConnection {
  /** From shape ID */
  fid: Ulid;
  /** To shape ID */
  tid: Ulid;
}

export interface DiagramCollaborator {
  /** ID */
  id: Ulid;
  /** ID */
  diagramId: Ulid;
  /** Username */
  userName: Username;
  permission: DiagramCollaboratorPermissionEnum;
  /** Datetime */
  createdAt: Datetime;
  /** Datetime */
  updatedAt: Datetime;
}

export interface DiagramComment {
  /** ID */
  id: Ulid;
  /** Username */
  userName: Username;
  text: string;
  /** Datetime */
  createdAt: Datetime;
}

export interface OnlineUser {
  /** ID */
  id: Ulid;
  /** Username */
  userName: Username;
}

export interface CursorPosition {
  /** ID */
  userId: Ulid;
  /** Username */
  userName: Username;
  /** @format double */
  x: number;
  /** @format double */
  y: number;
}

export interface EmbedInfo {
  /** ID */
  id: Ulid;
  embedCode: string;
  isActive: boolean;
  /** @format int32 */
  viewsCount: number;
  /** Datetime */
  lastViewedAt?: Datetime | null;
  /** Datetime */
  createdAt: Datetime;
  /** Datetime */
  updatedAt: Datetime;
}

import type {
  AxiosInstance,
  AxiosRequestConfig,
  HeadersDefaults,
  ResponseType,
} from "axios";
import axios from "axios";

export type QueryParamsType = Record<string | number, any>;

export interface FullRequestParams
  extends Omit<AxiosRequestConfig, "data" | "params" | "url" | "responseType"> {
  /** set parameter to `true` for call `securityWorker` for this request */
  secure?: boolean;
  /** request path */
  path: string;
  /** content type of request body */
  type?: ContentType;
  /** query params */
  query?: QueryParamsType;
  /** format of response (i.e. response.json() -> format: "json") */
  format?: ResponseType;
  /** request body */
  body?: unknown;
}

export type RequestParams = Omit<
  FullRequestParams,
  "body" | "method" | "query" | "path"
>;

export interface ApiConfig<SecurityDataType = unknown>
  extends Omit<AxiosRequestConfig, "data" | "cancelToken"> {
  securityWorker?: (
    securityData: SecurityDataType | null,
  ) => Promise<AxiosRequestConfig | void> | AxiosRequestConfig | void;
  secure?: boolean;
  format?: ResponseType;
}

export type ContentType =
  | "application/json"
  | "application/vnd.api+json"
  | "multipart/form-data"
  | "application/x-www-form-urlencoded"
  | "text/plain";

export class HttpClient<SecurityDataType = unknown> {
  public instance: AxiosInstance;
  private securityData: SecurityDataType | null = null;
  private securityWorker?: ApiConfig<SecurityDataType>["securityWorker"];
  private secure?: boolean;
  private format?: ResponseType;

  constructor({
    securityWorker,
    secure,
    format,
    ...axiosConfig
  }: ApiConfig<SecurityDataType> = {}) {
    this.instance = axios.create({
      ...axiosConfig,
      baseURL: axiosConfig.baseURL || "/",
    });
    this.secure = secure;
    this.format = format;
    this.securityWorker = securityWorker;
  }

  public setSecurityData = (data: SecurityDataType | null) => {
    this.securityData = data;
  };

  protected mergeRequestParams(
    params1: AxiosRequestConfig,
    params2?: AxiosRequestConfig,
  ): AxiosRequestConfig {
    const method = params1.method || (params2 && params2.method);

    return {
      ...this.instance.defaults,
      ...params1,
      ...(params2 || {}),
      headers: {
        ...((method &&
          this.instance.defaults.headers[
            method.toLowerCase() as keyof HeadersDefaults
          ]) ||
          {}),
        ...(params1.headers || {}),
        ...((params2 && params2.headers) || {}),
      },
    };
  }

  protected stringifyFormItem(formItem: unknown) {
    if (typeof formItem === "object" && formItem !== null) {
      return JSON.stringify(formItem);
    } else {
      return `${formItem}`;
    }
  }

  protected createFormData(input: Record<string, unknown>): FormData {
    if (input instanceof FormData) {
      return input;
    }
    return Object.keys(input || {}).reduce((formData, key) => {
      const property = input[key];
      const propertyContent: any[] =
        property instanceof Array ? property : [property];

      for (const formItem of propertyContent) {
        const isFileType = formItem instanceof Blob || formItem instanceof File;
        formData.append(
          key,
          isFileType ? formItem : this.stringifyFormItem(formItem),
        );
      }

      return formData;
    }, new FormData());
  }

  public request = async <T = any, _E = any>({
    secure,
    path,
    type,
    query,
    format,
    body,
    ...params
  }: FullRequestParams): Promise<T> => {
    const secureParams =
      ((typeof secure === "boolean" ? secure : this.secure) &&
        this.securityWorker &&
        (await this.securityWorker(this.securityData))) ||
      {};
    const requestParams = this.mergeRequestParams(params, secureParams);
    const responseFormat = format || this.format || undefined;

    if (
      type === "multipart/form-data" &&
      body &&
      body !== null &&
      typeof body === "object"
    ) {
      body = this.createFormData(body as Record<string, unknown>);
    }

    if (
      type === "text/plain" &&
      body &&
      body !== null &&
      typeof body !== "string"
    ) {
      body = JSON.stringify(body);
    }

    return this.instance
      .request({
        ...requestParams,
        headers: {
          ...(requestParams.headers || {}),
          ...(type ? { "Content-Type": type } : {}),
        },
        params: query,
        responseType: responseFormat,
        data: body,
        url: path,
      })
      .then((response) => response.data);
  };
}

/**
 * @title Klubok API
 * @version 0.1.0
 * @license AGPLv3 (https://www.gnu.org/licenses/agpl-3.0.html)
 * @baseUrl /
 */
export class ApiSpec<SecurityDataType extends unknown> {
  http: HttpClient<SecurityDataType>;

  constructor(http: HttpClient<SecurityDataType>) {
    this.http = http;
  }

  rootService = {
    /**
     * No description
     *
     * @tags RootService
     * @name Index
     * @summary Get index.
     * @request GET:/
     * @response `200` `IndexResponse` A successful response.
     * @response `429` `GeneralError`
     */
    index: (params: RequestParams = {}) =>
      this.http.request<IndexResponse, GeneralError>({
        path: `/`,
        method: "GET",
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags RootService
     * @name HealthLive
     * @summary Check liveness.
     * @request GET:/health/live
     * @response `204` `void` The process is alive.
     * @response `429` `GeneralError`
     * @response `500` `GeneralError`
     */
    healthLive: (params: RequestParams = {}) =>
      this.http.request<void, GeneralError>({
        path: `/health/live`,
        method: "GET",
        ...params,
      }),

    /**
     * No description
     *
     * @tags RootService
     * @name HealthReady
     * @summary Check readiness.
     * @request GET:/health/ready
     * @response `204` `void` The service is ready to accept traffic.
     * @response `429` `GeneralError`
     * @response `500` `GeneralError`
     * @response `503` `void` A critical dependency is unavailable.
     */
    healthReady: (params: RequestParams = {}) =>
      this.http.request<void, GeneralError | void>({
        path: `/health/ready`,
        method: "GET",
        ...params,
      }),

    /**
     * No description
     *
     * @tags RootService
     * @name Features
     * @summary Get feature flags.
     * @request GET:/features
     * @response `200` `FeaturesResponse` A successful response.
     * @response `429` `GeneralError`
     * @response `500` `GeneralError`
     */
    features: (params: RequestParams = {}) =>
      this.http.request<FeaturesResponse, GeneralError>({
        path: `/features`,
        method: "GET",
        format: "json",
        ...params,
      }),
  };
  diagramService = {
    /**
     * No description
     *
     * @tags DiagramService
     * @name DiagramList
     * @summary List diagrams.
     * @request GET:/diagrams
     * @secure
     * @response `200` `DiagramListResponse` A successful response.
     * @response `401` `GeneralError`
     * @response `429` `GeneralError`
     * @response `500` `GeneralError`
     */
    diagramList: (params: RequestParams = {}) =>
      this.http.request<DiagramListResponse, GeneralError>({
        path: `/diagrams`,
        method: "GET",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags DiagramService
     * @name DiagramCreate
     * @summary Create diagram.
     * @request POST:/diagrams
     * @secure
     * @response `201` `DiagramResponse` A successful response.
     * @response `400` `GeneralError`
     * @response `401` `GeneralError`
     * @response `429` `GeneralError`
     * @response `500` `GeneralError`
     */
    diagramCreate: (data: DiagramCreateRequest, params: RequestParams = {}) =>
      this.http.request<DiagramResponse, GeneralError>({
        path: `/diagrams`,
        method: "POST",
        body: data,
        secure: true,
        type: "application/json",
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags DiagramService
     * @name DiagramImport
     * @summary Import a diagram.
     * @request POST:/diagrams/import
     * @secure
     * @response `201` `DiagramResponse` A successful response.
     * @response `400` `GeneralError`
     * @response `401` `GeneralError`
     * @response `429` `GeneralError`
     * @response `500` `GeneralError`
     */
    diagramImport: (data: DiagramImportRequest, params: RequestParams = {}) =>
      this.http.request<DiagramResponse, GeneralError>({
        path: `/diagrams/import`,
        method: "POST",
        body: data,
        secure: true,
        type: "application/json",
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags DiagramService
     * @name DiagramSetTitle
     * @summary Set title diagram.
     * @request PATCH:/diagrams/{diagramId}/title
     * @secure
     * @response `204` `void` A successful response.
     * @response `400` `GeneralError`
     * @response `401` `GeneralError`
     * @response `403` `GeneralError`
     * @response `429` `GeneralError`
     * @response `500` `GeneralError`
     */
    diagramSetTitle: (
      diagramId: Ulid,
      data: DiagramSetTitleRequest,
      params: RequestParams = {},
    ) =>
      this.http.request<void, GeneralError>({
        path: `/diagrams/${diagramId}/title`,
        method: "PATCH",
        body: data,
        secure: true,
        type: "application/json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags DiagramService
     * @name DiagramInfo
     * @summary Info diagram.
     * @request GET:/diagrams/{diagramId}
     * @secure
     * @response `200` `DiagramInfoResponse` A successful response.
     * @response `401` `GeneralError`
     * @response `403` `GeneralError`
     * @response `429` `GeneralError`
     * @response `500` `GeneralError`
     */
    diagramInfo: (diagramId: Ulid, params: RequestParams = {}) =>
      this.http.request<DiagramInfoResponse, GeneralError>({
        path: `/diagrams/${diagramId}`,
        method: "GET",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags DiagramService
     * @name DiagramDelete
     * @summary Delete diagram.
     * @request DELETE:/diagrams/{diagramId}
     * @secure
     * @response `204` `void` A successful response.
     * @response `401` `GeneralError`
     * @response `403` `GeneralError`
     * @response `429` `GeneralError`
     * @response `500` `GeneralError`
     */
    diagramDelete: (diagramId: Ulid, params: RequestParams = {}) =>
      this.http.request<void, GeneralError>({
        path: `/diagrams/${diagramId}`,
        method: "DELETE",
        secure: true,
        ...params,
      }),

    /**
     * No description
     *
     * @tags DiagramService
     * @name DiagramStat
     * @summary Stat diagram.
     * @request GET:/diagrams/{diagramId}/stat
     * @secure
     * @response `200` `DiagramStatResponse` A successful response.
     * @response `401` `GeneralError`
     * @response `403` `GeneralError`
     * @response `429` `GeneralError`
     * @response `500` `GeneralError`
     */
    diagramStat: (diagramId: Ulid, params: RequestParams = {}) =>
      this.http.request<DiagramStatResponse, GeneralError>({
        path: `/diagrams/${diagramId}/stat`,
        method: "GET",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags DiagramService
     * @name DiagramOnlineUsers
     * @summary Online users on diagram.
     * @request GET:/diagrams/{diagramId}/onlineUsers
     * @secure
     * @response `200` `DiagramOnlineUsersResponse` A successful response.
     * @response `401` `GeneralError`
     * @response `403` `GeneralError`
     * @response `429` `GeneralError`
     * @response `500` `GeneralError`
     */
    diagramOnlineUsers: (diagramId: Ulid, params: RequestParams = {}) =>
      this.http.request<DiagramOnlineUsersResponse, GeneralError>({
        path: `/diagrams/${diagramId}/onlineUsers`,
        method: "GET",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags DiagramService
     * @name DiagramClone
     * @summary Clone a diagram with all its contents.
     * @request POST:/diagrams/{diagramId}/clone
     * @secure
     * @response `201` `DiagramResponse` A successful response.
     * @response `401` `GeneralError`
     * @response `403` `GeneralError`
     * @response `404` `GeneralError`
     * @response `429` `GeneralError`
     * @response `500` `GeneralError`
     */
    diagramClone: (diagramId: Ulid, params: RequestParams = {}) =>
      this.http.request<DiagramResponse, GeneralError>({
        path: `/diagrams/${diagramId}/clone`,
        method: "POST",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags DiagramService
     * @name CollaboratorList
     * @summary List collaborators on diagram.
     * @request GET:/diagrams/{diagramId}/collaborators
     * @secure
     * @response `200` `CollaboratorListResponse` A successful response.
     * @response `401` `GeneralError`
     * @response `403` `GeneralError`
     * @response `429` `GeneralError`
     * @response `500` `GeneralError`
     */
    collaboratorList: (diagramId: Ulid, params: RequestParams = {}) =>
      this.http.request<CollaboratorListResponse, GeneralError>({
        path: `/diagrams/${diagramId}/collaborators`,
        method: "GET",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags DiagramService
     * @name CollaboratorUpsert
     * @summary Upsert collaborator on diagram.
     * @request POST:/diagrams/{diagramId}/collaborators
     * @secure
     * @response `200` `CollaboratorUpsertResponse` A successful response.
     * @response `400` `GeneralError`
     * @response `401` `GeneralError`
     * @response `403` `GeneralError`
     * @response `404` `GeneralError`
     * @response `429` `GeneralError`
     * @response `500` `GeneralError`
     */
    collaboratorUpsert: (
      diagramId: Ulid,
      data: CollaboratorUpsertRequest,
      params: RequestParams = {},
    ) =>
      this.http.request<CollaboratorUpsertResponse, GeneralError>({
        path: `/diagrams/${diagramId}/collaborators`,
        method: "POST",
        body: data,
        secure: true,
        type: "application/json",
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags DiagramService
     * @name CollaboratorDelete
     * @summary Delete collaborator on diagram.
     * @request DELETE:/diagrams/{diagramId}/collaborators/{collaboratorId}
     * @secure
     * @response `204` `void` A successful response.
     * @response `401` `GeneralError`
     * @response `403` `GeneralError`
     * @response `429` `GeneralError`
     * @response `500` `GeneralError`
     */
    collaboratorDelete: (
      diagramId: Ulid,
      collaboratorId: Ulid,
      params: RequestParams = {},
    ) =>
      this.http.request<void, GeneralError>({
        path: `/diagrams/${diagramId}/collaborators/${collaboratorId}`,
        method: "DELETE",
        secure: true,
        ...params,
      }),

    /**
     * No description
     *
     * @tags DiagramService
     * @name ShapeUpsert
     * @summary Upsert shape on diagram.
     * @request POST:/diagrams/{diagramId}/shapes
     * @secure
     * @response `204` `void` A successful response.
     * @response `400` `GeneralError`
     * @response `401` `GeneralError`
     * @response `403` `GeneralError`
     * @response `429` `GeneralError`
     * @response `500` `GeneralError`
     */
    shapeUpsert: (
      diagramId: Ulid,
      data: ShapeUpsertRequest,
      params: RequestParams = {},
    ) =>
      this.http.request<void, GeneralError>({
        path: `/diagrams/${diagramId}/shapes`,
        method: "POST",
        body: data,
        secure: true,
        type: "application/json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags DiagramService
     * @name ShapeBatchUpsert
     * @summary Batch upsert shapes on diagram.
     * @request POST:/diagrams/{diagramId}/shapes/batch
     * @secure
     * @response `204` `void` A successful response.
     * @response `400` `GeneralError`
     * @response `401` `GeneralError`
     * @response `403` `GeneralError`
     * @response `429` `GeneralError`
     * @response `500` `GeneralError`
     */
    shapeBatchUpsert: (
      diagramId: Ulid,
      data: ShapeBatchUpsertRequest,
      params: RequestParams = {},
    ) =>
      this.http.request<void, GeneralError>({
        path: `/diagrams/${diagramId}/shapes/batch`,
        method: "POST",
        body: data,
        secure: true,
        type: "application/json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags DiagramService
     * @name ShapeDelete
     * @summary Delete shape on diagram.
     * @request DELETE:/diagrams/{diagramId}/shapes/{diagramShapeId}
     * @secure
     * @response `204` `void` A successful response.
     * @response `401` `GeneralError`
     * @response `403` `GeneralError`
     * @response `429` `GeneralError`
     * @response `500` `GeneralError`
     */
    shapeDelete: (
      diagramId: Ulid,
      diagramShapeId: Ulid,
      params: RequestParams = {},
    ) =>
      this.http.request<void, GeneralError>({
        path: `/diagrams/${diagramId}/shapes/${diagramShapeId}`,
        method: "DELETE",
        secure: true,
        ...params,
      }),

    /**
     * No description
     *
     * @tags DiagramService
     * @name ConnectionUpsert
     * @summary Upsert connection on diagram.
     * @request POST:/diagrams/{diagramId}/connections
     * @secure
     * @response `204` `void` A successful response.
     * @response `400` `GeneralError`
     * @response `401` `GeneralError`
     * @response `403` `GeneralError`
     * @response `429` `GeneralError`
     * @response `500` `GeneralError`
     */
    connectionUpsert: (
      diagramId: Ulid,
      data: ConnectionUpsertRequest,
      params: RequestParams = {},
    ) =>
      this.http.request<void, GeneralError>({
        path: `/diagrams/${diagramId}/connections`,
        method: "POST",
        body: data,
        secure: true,
        type: "application/json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags DiagramService
     * @name ConnectionBatchUpsert
     * @summary Batch upsert connections on diagram.
     * @request POST:/diagrams/{diagramId}/connections/batch
     * @secure
     * @response `204` `void` A successful response.
     * @response `400` `GeneralError`
     * @response `401` `GeneralError`
     * @response `403` `GeneralError`
     * @response `429` `GeneralError`
     * @response `500` `GeneralError`
     */
    connectionBatchUpsert: (
      diagramId: Ulid,
      data: ConnectionBatchUpsertRequest,
      params: RequestParams = {},
    ) =>
      this.http.request<void, GeneralError>({
        path: `/diagrams/${diagramId}/connections/batch`,
        method: "POST",
        body: data,
        secure: true,
        type: "application/json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags DiagramService
     * @name ConnectionDelete
     * @summary Delete connection on diagram.
     * @request DELETE:/diagrams/{diagramId}/connections/{fromShapeId}/{toShapeId}
     * @secure
     * @response `204` `void` A successful response.
     * @response `401` `GeneralError`
     * @response `403` `GeneralError`
     * @response `429` `GeneralError`
     * @response `500` `GeneralError`
     */
    connectionDelete: (
      diagramId: Ulid,
      fromShapeId: Ulid,
      toShapeId: Ulid,
      params: RequestParams = {},
    ) =>
      this.http.request<void, GeneralError>({
        path: `/diagrams/${diagramId}/connections/${fromShapeId}/${toShapeId}`,
        method: "DELETE",
        secure: true,
        ...params,
      }),

    /**
     * No description
     *
     * @tags DiagramService
     * @name CommentList
     * @summary List comments on diagram.
     * @request GET:/diagrams/{diagramId}/comments/{diagramShapeId}
     * @secure
     * @response `200` `CommentListResponse` A successful response.
     * @response `401` `GeneralError`
     * @response `403` `GeneralError`
     * @response `429` `GeneralError`
     * @response `500` `GeneralError`
     */
    commentList: (
      diagramId: Ulid,
      diagramShapeId: Ulid,
      params: RequestParams = {},
    ) =>
      this.http.request<CommentListResponse, GeneralError>({
        path: `/diagrams/${diagramId}/comments/${diagramShapeId}`,
        method: "GET",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags DiagramService
     * @name CommentCreate
     * @summary Create comment on diagram.
     * @request POST:/diagrams/{diagramId}/comments/{diagramShapeId}
     * @secure
     * @response `201` `CommentCreateResponse` A successful response.
     * @response `400` `GeneralError`
     * @response `401` `GeneralError`
     * @response `403` `GeneralError`
     * @response `429` `GeneralError`
     * @response `500` `GeneralError`
     */
    commentCreate: (
      diagramId: Ulid,
      diagramShapeId: Ulid,
      data: CommentCreateRequest,
      params: RequestParams = {},
    ) =>
      this.http.request<CommentCreateResponse, GeneralError>({
        path: `/diagrams/${diagramId}/comments/${diagramShapeId}`,
        method: "POST",
        body: data,
        secure: true,
        type: "application/json",
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags DiagramService
     * @name EmbedCreate
     * @summary Create embed code for diagram.
     * @request POST:/diagrams/{diagramId}/embeds
     * @secure
     * @response `201` `EmbedInfoResponse` A successful response.
     * @response `400` `GeneralError`
     * @response `401` `GeneralError`
     * @response `403` `GeneralError`
     * @response `429` `GeneralError`
     * @response `500` `GeneralError`
     */
    embedCreate: (diagramId: Ulid, params: RequestParams = {}) =>
      this.http.request<EmbedInfoResponse, GeneralError>({
        path: `/diagrams/${diagramId}/embeds`,
        method: "POST",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags DiagramService
     * @name EmbedList
     * @summary List embed codes for diagram.
     * @request GET:/diagrams/{diagramId}/embeds
     * @secure
     * @response `200` `EmbedListResponse` A successful response.
     * @response `401` `GeneralError`
     * @response `403` `GeneralError`
     * @response `429` `GeneralError`
     * @response `500` `GeneralError`
     */
    embedList: (diagramId: Ulid, params: RequestParams = {}) =>
      this.http.request<EmbedListResponse, GeneralError>({
        path: `/diagrams/${diagramId}/embeds`,
        method: "GET",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags DiagramService
     * @name EmbedGet
     * @summary Get embed code info.
     * @request GET:/diagrams/{diagramId}/embeds/{embedId}
     * @secure
     * @response `200` `EmbedInfoResponse` A successful response.
     * @response `401` `GeneralError`
     * @response `403` `GeneralError`
     * @response `404` `GeneralError`
     * @response `429` `GeneralError`
     * @response `500` `GeneralError`
     */
    embedGet: (diagramId: Ulid, embedId: Ulid, params: RequestParams = {}) =>
      this.http.request<EmbedInfoResponse, GeneralError>({
        path: `/diagrams/${diagramId}/embeds/${embedId}`,
        method: "GET",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags DiagramService
     * @name EmbedDelete
     * @summary Revoke embed code (soft-delete).
     * @request DELETE:/diagrams/{diagramId}/embeds/{embedId}
     * @secure
     * @response `204` `void` A successful response.
     * @response `401` `GeneralError`
     * @response `403` `GeneralError`
     * @response `404` `GeneralError`
     * @response `429` `GeneralError`
     * @response `500` `GeneralError`
     */
    embedDelete: (diagramId: Ulid, embedId: Ulid, params: RequestParams = {}) =>
      this.http.request<void, GeneralError>({
        path: `/diagrams/${diagramId}/embeds/${embedId}`,
        method: "DELETE",
        secure: true,
        ...params,
      }),

    /**
     * No description
     *
     * @tags DiagramService
     * @name EmbedData
     * @summary Get embed data by public token for iframe.
     * @request GET:/embed/{token}
     * @response `200` `EmbedDataResponse` A successful response.
     * @response `404` `GeneralError`
     * @response `429` `GeneralError`
     * @response `500` `GeneralError`
     */
    embedData: (token: string, params: RequestParams = {}) =>
      this.http.request<EmbedDataResponse, GeneralError>({
        path: `/embed/${token}`,
        method: "GET",
        format: "json",
        ...params,
      }),
  };
  userService = {
    /**
     * No description
     *
     * @tags UserService
     * @name UserRegister
     * @summary Register user.
     * @request POST:/users/register
     * @response `200` `UserRegisterResponse` A successful response.
     * @response `400` `GeneralError`
     * @response `403` `GeneralError`
     * @response `409` `GeneralError`
     * @response `429` `GeneralError`
     * @response `500` `GeneralError`
     */
    userRegister: (data: UserRegisterRequest, params: RequestParams = {}) =>
      this.http.request<UserRegisterResponse, GeneralError>({
        path: `/users/register`,
        method: "POST",
        body: data,
        type: "application/json",
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags UserService
     * @name UserActivate
     * @summary Activate user.
     * @request POST:/users/activate
     * @response `204` `void` A successful response.
     * @response `400` `GeneralError`
     * @response `403` `GeneralError`
     * @response `404` `GeneralError`
     * @response `429` `GeneralError`
     * @response `500` `GeneralError`
     */
    userActivate: (data: UserActivateRequest, params: RequestParams = {}) =>
      this.http.request<void, GeneralError>({
        path: `/users/activate`,
        method: "POST",
        body: data,
        type: "application/json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags UserService
     * @name UserLogin
     * @summary Login user.
     * @request POST:/users/login
     * @response `204` `void` A successful response.
     * @response `400` `GeneralError`
     * @response `403` `GeneralError`
     * @response `404` `GeneralError`
     * @response `429` `GeneralError`
     * @response `500` `GeneralError`
     */
    userLogin: (data: UserLoginRequest, params: RequestParams = {}) =>
      this.http.request<void, GeneralError>({
        path: `/users/login`,
        method: "POST",
        body: data,
        type: "application/json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags UserService
     * @name UserLoginVerify
     * @summary Login verify user.
     * @request POST:/users/login/verify
     * @response `204` `void` A successful response.
     * @response `400` `GeneralError`
     * @response `403` `GeneralError`
     * @response `404` `GeneralError`
     * @response `429` `GeneralError`
     * @response `500` `GeneralError`
     */
    userLoginVerify: (
      data: UserLoginVerifyRequest,
      params: RequestParams = {},
    ) =>
      this.http.request<void, GeneralError>({
        path: `/users/login/verify`,
        method: "POST",
        body: data,
        type: "application/json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags UserService
     * @name UserLogout
     * @summary Logout current user.
     * @request DELETE:/users/logout
     * @secure
     * @response `204` `void` A successful response.
     * @response `401` `GeneralError`
     * @response `403` `GeneralError`
     * @response `429` `GeneralError`
     * @response `500` `GeneralError`
     */
    userLogout: (params: RequestParams = {}) =>
      this.http.request<void, GeneralError>({
        path: `/users/logout`,
        method: "DELETE",
        secure: true,
        ...params,
      }),

    /**
     * No description
     *
     * @tags UserService
     * @name UserFake
     * @summary Create fake user.
     * @request POST:/users/fake
     * @response `200` `UserMeResponse` A successful response.
     * @response `404` `GeneralError`
     * @response `429` `GeneralError`
     * @response `500` `GeneralError`
     */
    userFake: (params: RequestParams = {}) =>
      this.http.request<UserMeResponse, GeneralError>({
        path: `/users/fake`,
        method: "POST",
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags UserService
     * @name SsoLogin
     * @summary Start SSO login.
     * @request GET:/auth/sso/login
     * @response `302` `void` Redirect to OIDC provider.
     * @response `404` `GeneralError`
     * @response `429` `GeneralError`
     * @response `500` `GeneralError`
     */
    ssoLogin: (params: RequestParams = {}) =>
      this.http.request<any, void | GeneralError>({
        path: `/auth/sso/login`,
        method: "GET",
        ...params,
      }),

    /**
     * No description
     *
     * @tags UserService
     * @name SsoCallback
     * @summary SSO callback from provider.
     * @request GET:/auth/sso/callback
     * @response `302` `void` Redirect to app with session cookie.
     * @response `400` `GeneralError`
     * @response `404` `GeneralError`
     * @response `429` `GeneralError`
     * @response `500` `GeneralError`
     */
    ssoCallback: (
      query: {
        /** @minLength 1 */
        code: string;
        /** @minLength 1 */
        state: string;
      },
      params: RequestParams = {},
    ) =>
      this.http.request<any, void | GeneralError>({
        path: `/auth/sso/callback`,
        method: "GET",
        query: query,
        ...params,
      }),

    /**
     * No description
     *
     * @tags UserService
     * @name UserMe
     * @summary Get current user.
     * @request GET:/users/me
     * @secure
     * @response `200` `UserMeResponse` A successful response.
     * @response `401` `GeneralError`
     * @response `403` `GeneralError`
     * @response `429` `GeneralError`
     * @response `500` `GeneralError`
     */
    userMe: (params: RequestParams = {}) =>
      this.http.request<UserMeResponse, GeneralError>({
        path: `/users/me`,
        method: "GET",
        secure: true,
        format: "json",
        ...params,
      }),
  };
}
