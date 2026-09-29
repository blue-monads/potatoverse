// --- Example Beta Plugin Loader Script ---
(function() {
  console.log("[Example Beta Plugin] Initializing inside host space...");

  const pluginMeta = {
    name: "Example Beta Plugin",
    key: "pexample-beta",
    version: "0.0.1"
  };

  // Register with host app registry if available
  if (window.AlphaApp && typeof window.AlphaApp.registerPlugin === "function") {
    window.AlphaApp.registerPlugin(pluginMeta);
  }

  // Inject plugin card into host UI
  function renderPluginWidget() {
    const slots = document.getElementById("plugin-slots");
    if (!slots) return;

    const noPlugins = document.getElementById("no-plugins-notice");
    if (noPlugins) noPlugins.style.display = "none";

    const widget = document.createElement("div");
    widget.id = "plugin-widget-beta";
    widget.style.cssText = `
      border: 1px solid #c7d2fe;
      background: #eef2ff;
      border-radius: 8px;
      padding: 16px;
      margin-top: 12px;
    `;

    widget.innerHTML = `
      <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 8px;">
        <strong style="color: #3730a3; font-size: 15px;">🔌 Example Beta Plugin</strong>
        <span style="font-size: 11px; background: #c7d2fe; color: #312e81; padding: 2px 8px; border-radius: 12px;">Active</span>
      </div>
      <p style="margin: 0 0 12px 0; font-size: 13px; color: #4338ca;">
        Loaded dynamically via <code>/zz/core/space/:space_key/plugin_loaders.js</code>.
      </p>
      <div style="display: flex; gap: 8px;">
        <button id="beta-ping-btn" style="background: #4f46e5; color: white; border: none; padding: 6px 14px; border-radius: 4px; font-size: 13px; cursor: pointer;">
          Call Plugin API (/ping)
        </button>
        <button id="beta-info-btn" style="background: #6366f1; color: white; border: none; padding: 6px 14px; border-radius: 4px; font-size: 13px; cursor: pointer;">
          Get Plugin Info
        </button>
      </div>
      <pre id="beta-output" style="margin-top: 10px; background: #1e1b4b; color: #a5b4fc; padding: 10px; border-radius: 6px; font-size: 12px; display: none; overflow-x: auto;"></pre>
    `;

    slots.appendChild(widget);

    // Helper to determine the host space key from current path or host registry
    const hostSpaceKey = (window.AlphaApp && window.AlphaApp.spaceKey) || "pexample-alpha";

    document.getElementById("beta-ping-btn").addEventListener("click", async () => {
      const out = document.getElementById("beta-output");
      out.style.display = "block";
      out.textContent = "Calling /zz/api/plugin/" + hostSpaceKey + "/pexample-beta/ping ...";

      try {
        const res = await fetch("/zz/api/plugin/" + hostSpaceKey + "/pexample-beta/ping");
        const json = await res.json();
        out.textContent = JSON.stringify(json, null, 2);
      } catch (e) {
        out.textContent = "Request failed: " + e.message;
      }
    });

    document.getElementById("beta-info-btn").addEventListener("click", async () => {
      const out = document.getElementById("beta-output");
      out.style.display = "block";
      out.textContent = "Calling /zz/api/plugin/" + hostSpaceKey + "/pexample-beta/info ...";

      try {
        const res = await fetch("/zz/api/plugin/" + hostSpaceKey + "/pexample-beta/info");
        const json = await res.json();
        out.textContent = JSON.stringify(json, null, 2);
      } catch (e) {
        out.textContent = "Request failed: " + e.message;
      }
    });
  }

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", renderPluginWidget);
  } else {
    renderPluginWidget();
  }
})();
