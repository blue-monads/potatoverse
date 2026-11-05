window.Cloudy = {
  tokenKey: "cloudy_token",
  pages: "/zz/cloudy/pages",

  token() {
    return localStorage.getItem(this.tokenKey);
  },

  setToken(t) {
    localStorage.setItem(this.tokenKey, t);
  },

  page(name) {
    return this.pages + "/" + name;
  },

  signOut() {
    localStorage.removeItem(this.tokenKey);
    location.href = this.page("login.html");
  },

  headers(json) {
    const h = { Authorization: "Bearer " + this.token() };
    if (json) h["Content-Type"] = "application/json";
    return h;
  },

  async api(path, opts) {
    const res = await fetch(path, opts);
    const data = await res.json().catch(() => ({}));
    if (!res.ok) throw new Error(data.error || res.statusText || "request failed");
    return data;
  },

  async download(path, filename) {
    const res = await fetch(path, { headers: this.headers() });
    if (!res.ok) {
      const data = await res.json().catch(() => ({}));
      throw new Error(data.error || res.statusText || "download failed");
    }
    const blob = await res.blob();
    const url = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url;
    a.download = filename;
    document.body.appendChild(a);
    a.click();
    a.remove();
    URL.revokeObjectURL(url);
    return filename;
  },

  exportFilename(tenant) {
    const d = new Date();
    const y = d.getFullYear();
    const m = String(d.getMonth() + 1).padStart(2, "0");
    const day = String(d.getDate()).padStart(2, "0");
    return tenant + "-app-" + y + "-" + m + "-" + day + ".db";
  },

  afterLogin(user, token) {
    this.setToken(token);
    location.href = this.page(user && user.utype === "admin" ? "admin/portal.html" : "portal.html");
  },

  tenantURL(session, tenant) {
    const host = tenant + "." + session.domain;
    return "http://" + host + ":" + session.port;
  },

  async requireSession(opts) {
    opts = opts || {};
    if (!this.token()) {
      location.href = this.page("login.html");
      return null;
    }
    try {
      const data = await this.api("/zz/cloudy/me", { headers: this.headers() });
      const utype = data.user && data.user.utype;
      if (opts.admin && utype !== "admin") {
        location.href = this.page("portal.html");
        return null;
      }
      return data;
    } catch (err) {
      this.signOut();
      return null;
    }
  },
};
