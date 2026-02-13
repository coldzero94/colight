import "@typespec/http";

using TypeSpec.Http;

namespace Colight.Common;

model ErrorDetail {
  message: string;
  code: string;
}

model ErrorResponse {
  @statusCode statusCode: 400 | 401 | 403 | 404 | 409 | 500;
  @body body: {
    error: ErrorDetail;
  };
}
