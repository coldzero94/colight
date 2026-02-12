import { defineConfig } from "@hey-api/openapi-ts";

export default defineConfig({
  input: "../../packages/protocol/tsp-output/openapi.yaml",
  output: "src/api/generated",
  plugins: [
    "@hey-api/typescript",
    "@hey-api/client-axios",
  ],
});
