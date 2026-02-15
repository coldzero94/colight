import "@typespec/http";
import "@typespec/rest";
import "@typespec/openapi3";
import "./src/common/errors.tsp";
import "./src/auth/auth.tsp";
import "./src/admin/admin.tsp";
import "./src/experience/experience.tsp";
import "./src/interview/interview.tsp";
import "./src/company/company.tsp";
import "./src/coaching/coaching.tsp";
import "./src/system/system.tsp";

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
