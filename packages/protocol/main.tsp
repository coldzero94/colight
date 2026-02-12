import "@typespec/http";
import "@typespec/rest";
import "@typespec/openapi3";

using Http;
using Rest;

@service(#{
  title: "Colight API"
})
@server("http://localhost:9000", "Local development")
@server("https://colight-api.koyeb.app", "Production")
namespace Colight;

// Health check
@route("/health")
@tag("System")
interface Health {
  @get op check(): { status: string };
}
