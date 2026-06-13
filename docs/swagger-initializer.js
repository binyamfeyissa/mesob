// Swagger UI initializer — mounted into the swagger-ui container at
// /usr/share/nginx/html/swagger-initializer.js
// Loads the single combined spec that covers all 14 Mesob services.
// Each operation carries its own path-level server URL so "Try it out"
// calls the correct service port directly (e.g. identity → :8001).
window.onload = function () {
  window.ui = SwaggerUIBundle({
    url: "specs/combined.yaml",
    dom_id: "#swagger-ui",
    deepLinking: true,
    displayRequestDuration: true,
    filter: true,
    tryItOutEnabled: true,
    persistAuthorization: true,
    presets: [
      SwaggerUIBundle.presets.apis,
      SwaggerUIStandalonePreset,
    ],
    plugins: [SwaggerUIBundle.plugins.DownloadUrl],
    layout: "StandaloneLayout",
  });
};
