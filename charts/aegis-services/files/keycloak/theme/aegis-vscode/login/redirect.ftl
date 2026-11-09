<#ftl output_format="html" encoding="UTF-8">
<#import "template.ftl" as layout>
<@layout.registrationLayout displayMessage=false displayInfo=false; section>
  <#if section = "header">
    ${msg("redirectTitle")}
  <#elseif section = "form">
    <div class="pf-c-alert pf-m-success" role="status" style="margin-bottom: var(--pf-global--spacer--lg);">
      <div class="pf-c-alert__icon">
        <span class="pf-icon pf-icon-ok"></span>
      </div>
      <p class="pf-c-alert__title">
        ${msg("redirectText")}
      </p>
    </div>
    <p style="margin-bottom: var(--pf-global--spacer--md);">
      <a class="pf-c-button pf-m-primary" href="${kcSanitize(url.redirectUri)}">${msg("redirectContinue")}</a>
    </p>
    <p id="kc-vscode-close-hint" style="display:none; margin-bottom: var(--pf-global--spacer--md);">
      ${msg("aegisCloseTabHint", "VS Code")}
    </p>
    <p style="color: var(--pf-global--Color--200); font-size: var(--pf-global--FontSize--sm);">
      ${msg("aegisRedirectHelper")}
    </p>
    <script>
      (function () {
        var redirectUrl = "${kcSanitize(url.redirectUri)}";
        var closeHint = document.getElementById("kc-vscode-close-hint");

        function revealHint() {
          if (closeHint) {
            closeHint.style.display = "block";
          }
        }

        function tryCloseWindow() {
          try {
            window.close();
          } catch (err) {
            // ignored – browsers will block close() on non-scripted windows
          }
          revealHint();
        }

        // Maintain the default auto-redirect behaviour.
        if (redirectUrl) {
          window.location.replace(redirectUrl);
        }

        // Attempt to close the tab shortly after signalling the client.
        window.setTimeout(tryCloseWindow, 1500);
      }());
    </script>
  </#if>
</@layout.registrationLayout>
