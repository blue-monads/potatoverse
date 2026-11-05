package cloudy

import (
	"context"
	"fmt"
	"io/fs"
	"log"
	"net"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	xutils "github.com/blue-monads/potatoverse/backend/utils"
	"github.com/gin-gonic/gin"
)

var tenantSlugRe = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?$`)

var reservedTenants = map[string]bool{
	"www":    true,
	"api":    true,
	"main":   true,
	"cloudy": true,
	"portal": true,
	"admin":  true,
	"static": true,
}

type signUpRequest struct {
	TeamName string `json:"team_name"`
	Tenant   string `json:"tenant"`
	Slug     string `json:"slug"`
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type loginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type addUserRequest struct {
	Tenant   string `json:"tenant"`
	Slug     string `json:"slug"`
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
	UType    string `json:"utype"`
	Verified bool   `json:"verified"`
}

type resetPasswordRequest struct {
	Password string `json:"password" binding:"required"`
}

type setDisabledRequest struct {
	IsDisabled bool `json:"is_disabled"`
}

func (a *CloudyApp) registerBaseRouter(router *gin.Engine) {
	r := router.Group("/zz/cloudy")

	// public
	r.GET("/", a.handleIndex)
	r.GET("/health", a.handleHealth)
	r.GET("/sign-up", a.handleSignUpInfo)
	r.POST("/sign-up", a.handleSignUp)
	r.GET("/verify", a.handleVerify)
	r.POST("/login", a.handleLogin)

	// authed
	authed := r.Group("/", a.authMiddleware())
	authed.GET("/me", a.handleMe)
	authed.GET("/my-teams", a.handleMyTeams)
	authed.GET("/teams/:id", a.handleGetTeam)
	authed.POST("/teams/:id/instances", a.handleCreateInstance)
	authed.POST("/apps/:name/load", a.loadApp)
	authed.GET("/apps/:name/export", a.exportApp)
	authed.GET("/apps/:name/open", a.redirectToApp)
	authed.GET("/sub-users", a.handleListSubUsers)
	authed.POST("/sub-users", a.handleAddSubUser)
	authed.POST("/sub-users/:id/reset-password", a.handleResetSubUserPassword)
	authed.POST("/sub-users/:id/disable", a.handleSetSubUserDisabled)

	// admin
	admin := authed.Group("/", a.adminMiddleware())
	admin.GET("/users", a.handleListUsers)
	admin.GET("/teams", a.handleListTeams)
	admin.POST("/users", a.handleAddUser)
	admin.POST("/users/:id/reset-password", a.handleResetPassword)
	admin.POST("/users/:id/disable", a.handleSetDisabled)
	admin.POST("/apps/:name/unload", a.unloadApp)

	r.GET("/pages", a.servePages)
	r.GET("/pages/*filepath", a.servePages)
}

func (a *CloudyApp) servePages(c *gin.Context) {
	rel := strings.TrimPrefix(path.Clean("/"+strings.TrimPrefix(c.Param("filepath"), "/")), "/")
	if rel == "" || rel == "." {
		c.Redirect(http.StatusFound, "/zz/cloudy/pages/login.html")
		return
	}

	fullPath := path.Join("pages", rel)
	if fullPath != "pages" && !strings.HasPrefix(fullPath, "pages/") {
		c.Status(http.StatusNotFound)
		return
	}

	f, err := fs.Stat(pageFiles, fullPath)
	if (err != nil || f.IsDir()) && path.Ext(rel) == "" {
		fullPath = path.Join("pages", rel+".html")
		f, err = fs.Stat(pageFiles, fullPath)
	}
	if err != nil || f.IsDir() {
		c.Status(http.StatusNotFound)
		return
	}

	c.FileFromFS(fullPath, http.FS(pageFiles))
}

func (a *CloudyApp) handleIndex(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"service": "cloudy",
		"domain":  normalizeDomain(a.config.Domain),
		"message": "tenant edge for potatoverse",
	})
}

func (a *CloudyApp) handleHealth(c *gin.Context) {
	users, err := a.listUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": err.Error()})
		return
	}

	a.mu.RLock()
	loaded := len(a.subApps)
	a.mu.RUnlock()

	c.JSON(http.StatusOK, gin.H{
		"ok":             true,
		"users":          len(users),
		"tenants_loaded": loaded,
	})
}

func (a *CloudyApp) handleSignUpInfo(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"fields":      []string{"slug", "name", "email", "password"},
		"tenant_host": fmt.Sprintf("<slug>.%s", normalizeDomain(a.config.Domain)),
	})
}

func (a *CloudyApp) handleSignUp(c *gin.Context) {
	var req signUpRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	teamName := strings.TrimSpace(req.TeamName)
	if teamName == "" {
		teamName = strings.TrimSpace(req.Slug)
	}
	if teamName == "" {
		teamName = strings.TrimSpace(req.Tenant)
	}
	if teamName == "" {
		teamName = req.Name + "'s Team"
	}

	passwordHash, err := hashSignupPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	user, team, err := a.insertUserWithTeam(req.Name, req.Email, passwordHash, teamName, UTypeNormal, false)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := a.sendVerificationEmail(user, team.Name); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "user created but failed to send verification email: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"user":    user,
		"team":    team,
		"message": "check your email to verify before logging in",
	})
}

func (a *CloudyApp) handleLogin(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := a.getUserByEmail(strings.TrimSpace(req.Email))
	if err != nil || user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
		return
	}
	if !xutils.VerifyPassword(user.Password, req.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
		return
	}
	if user.IsDisabled {
		c.JSON(http.StatusForbidden, gin.H{"error": "account is disabled"})
		return
	}
	if !user.IsVerified {
		c.JSON(http.StatusForbidden, gin.H{"error": "email not verified"})
		return
	}

	slug := ""
	inst, _, err := a.getPrimaryInstanceForUser(user.ID)
	if err == nil && inst != nil {
		slug = inst.Slug
	}

	token, err := a.encodeClaim(&Claim{
		UserID:  user.ID,
		Email:   user.Email,
		UType:   user.UType,
		Slug:    slug,
		Purpose: purposeAccess,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": token,
		"user":  user,
		"slug":  slug,
	})
}

func (a *CloudyApp) handleMe(c *gin.Context) {
	user := getUser(c)
	claim := getClaim(c)
	slug := claim.Slug
	if slug == "" {
		inst, _, err := a.getPrimaryInstanceForUser(user.ID)
		if err == nil && inst != nil {
			slug = inst.Slug
		}
	}
	domain := normalizeDomain(a.config.Domain)
	host := ""
	url := ""
	if slug != "" {
		host = fmt.Sprintf("%s.%s", slug, domain)
		url = fmt.Sprintf("http://%s:%d", host, a.config.Port)
	}
	c.JSON(http.StatusOK, gin.H{
		"user":   user,
		"claim":  claim,
		"slug":   slug,
		"domain": domain,
		"port":   a.config.Port,
		"host":   host,
		"url":    url,
	})
}

func (a *CloudyApp) handleVerify(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		a.renderVerify(c, http.StatusBadRequest, verifyPage{
			Title:   "Verification failed",
			Message: "This link is missing a token. Use the link from your email.",
		})
		return
	}

	claim, err := a.decodeClaim(token)
	if err != nil || claim.Purpose != purposeVerify {
		a.renderVerify(c, http.StatusBadRequest, verifyPage{
			Title:   "Verification failed",
			Message: "This link is invalid or has expired. Sign up again or request a new email.",
		})
		return
	}

	user, err := a.resolveClaimUser(claim)
	if err != nil || user == nil {
		a.renderVerify(c, http.StatusNotFound, verifyPage{
			Title:   "Verification failed",
			Message: "We could not find this account.",
		})
		return
	}

	if !user.IsVerified {
		if err := a.markUserVerified(user.ID); err != nil {
			a.renderVerify(c, http.StatusInternalServerError, verifyPage{
				Title:   "Verification failed",
				Message: "Could not mark this account as verified. Try again later.",
			})
			return
		}
		user.IsVerified = true
	}

	slug := claim.Slug
	if slug == "" {
		inst, _, _ := a.getPrimaryInstanceForUser(user.ID)
		if inst != nil {
			slug = inst.Slug
		}
	}

	if slug == "" {
		a.renderVerify(c, http.StatusOK, verifyPage{
			OK:      true,
			Title:   "You're verified",
			Message: "Your email is confirmed. You can now sign in to your team portal.",
		})
		return
	}

	host := fmt.Sprintf("%s.%s", slug, normalizeDomain(a.config.Domain))
	tenantURL := fmt.Sprintf("http://%s:%d", host, a.config.Port)

	a.mu.RLock()
	existing, alreadyLoaded := a.subApps[slug]
	a.mu.RUnlock()
	if alreadyLoaded {
		_ = existing.WaitReady(30 * time.Second)
		a.renderVerify(c, http.StatusOK, verifyPage{
			OK:      true,
			Title:   "You're verified",
			Message: "Your email is confirmed. Sign in to Cloudy or open your tenant app.",
			Tenant:  slug,
			Host:    host,
			URL:     tenantURL,
		})
		return
	}

	adminPass, err := xutils.GenerateRandomString(20)
	if err != nil {
		a.renderVerify(c, http.StatusInternalServerError, verifyPage{
			Title:   "Verification failed",
			Message: "Could not finish setting up your tenant. Try again later.",
		})
		return
	}

	_, err = a.provisionTenant(user, slug, adminPass)
	if err != nil {
		a.renderVerify(c, http.StatusInternalServerError, verifyPage{
			Title:   "Verification failed",
			Message: "Your email is verified, but the tenant app could not start: " + err.Error(),
			Tenant:  slug,
		})
		return
	}

	_ = a.sendMail(user.Email, "Your Cloudy instance is ready",
		fmt.Sprintf("Hi %s,\n\nInstance %s is ready.\nURL: %s\nAdmin user: %s\nAdmin password: %s\n",
			user.Fullname, slug, tenantURL, user.Fullname, adminPass),
		fmt.Sprintf(`<p>Hi %s,</p><p>Instance <strong>%s</strong> is ready.</p><p>URL: <a href="%s">%s</a></p><p>Admin user: %s<br/>Admin password: <code>%s</code></p>`,
			user.Fullname, slug, tenantURL, tenantURL, user.Fullname, adminPass),
	)

	a.renderVerify(c, http.StatusOK, verifyPage{
		OK:            true,
		Title:         "You're verified",
		Message:       "Your email is confirmed and your tenant app is ready.",
		Tenant:        slug,
		Host:          host,
		URL:           tenantURL,
		AdminName:     user.Fullname,
		AdminPassword: adminPass,
	})
}

func (a *CloudyApp) handleListUsers(c *gin.Context) {
	users, err := a.listUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	out := make([]gin.H, 0, len(users))
	for _, u := range users {
		inst, team, _ := a.getPrimaryInstanceForUser(u.ID)
		slug := ""
		teamName := ""
		loaded := false
		if inst != nil {
			slug = inst.Slug
			a.mu.RLock()
			_, loaded = a.subApps[slug]
			a.mu.RUnlock()
		}
		if team != nil {
			teamName = team.Name
		}
		out = append(out, gin.H{
			"id":          u.ID,
			"fullname":    u.Fullname,
			"email":       u.Email,
			"utype":       u.UType,
			"is_verified": u.IsVerified,
			"is_disabled": u.IsDisabled,
			"slug":        slug,
			"team_name":   teamName,
			"loaded":      loaded,
			"created_at":  u.CreatedAt,
			"updated_at":  u.UpdatedAt,
		})
	}
	c.JSON(http.StatusOK, gin.H{"users": out})
}

func (a *CloudyApp) handleListTeams(c *gin.Context) {
	teams, err := a.listTeams()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	insts, err := a.listPotatoInstances()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"teams":     teams,
		"instances": insts,
	})
}

func (a *CloudyApp) handleMyTeams(c *gin.Context) {
	u := getUser(c)
	if u == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	teams, err := a.getTeamsForUser(u.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"teams": teams})
}

func (a *CloudyApp) handleGetTeam(c *gin.Context) {
	u := getUser(c)
	claim := getClaim(c)
	if u == nil || claim == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	id, err := parseIDParam(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid team id"})
		return
	}

	if claim.UType != UTypeAdmin {
		inTeam, err := a.isUserInTeam(u.ID, id)
		if err != nil || !inTeam {
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
	}

	team, err := a.getTeamByID(id)
	if err != nil || team == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "team not found"})
		return
	}

	insts, err := a.getPotatoInstancesByTeamID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	a.mu.RLock()
	defer a.mu.RUnlock()

	outInsts := make([]gin.H, 0, len(insts))
	for _, inst := range insts {
		sub, loaded := a.subApps[inst.Slug]
		ready := false
		if sub != nil {
			ready = sub.IsReady()
		}
		outInsts = append(outInsts, gin.H{
			"id":          inst.ID,
			"slug":        inst.Slug,
			"description": inst.Description,
			"team_id":     inst.TeamID,
			"loaded":      loaded,
			"ready":       ready,
			"created_at":  inst.CreatedAt,
			"updated_at":  inst.UpdatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"team":      team,
		"instances": outInsts,
	})
}

type createInstanceRequest struct {
	Slug        string `json:"slug" binding:"required"`
	Description string `json:"description"`
}

func (a *CloudyApp) handleCreateInstance(c *gin.Context) {
	u := getUser(c)
	claim := getClaim(c)
	if u == nil || claim == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	teamID, err := parseIDParam(c.Param("id"))
	if err != nil || teamID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid team id"})
		return
	}

	if claim.UType != UTypeAdmin {
		inTeam, err := a.isUserInTeam(u.ID, teamID)
		if err != nil || !inTeam {
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
	}

	var req createInstanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	slug := strings.ToLower(strings.TrimSpace(req.Slug))
	if err := validateTenantSlug(slug); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if a.slugExists(slug) {
		c.JSON(http.StatusConflict, gin.H{"error": "instance slug already exists"})
		return
	}

	inst, err := a.createPotatoInstance(teamID, slug, req.Description)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"ok":       true,
		"instance": inst,
	})
}

func (a *CloudyApp) handleAddUser(c *gin.Context) {
	var req addUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	teamName := strings.TrimSpace(req.Slug)
	if teamName == "" {
		teamName = strings.TrimSpace(req.Tenant)
	}
	if teamName == "" {
		teamName = req.Name + "'s Team"
	}

	utype := req.UType
	if utype == "" {
		utype = UTypeNormal
	}
	if !validUType(utype) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "utype must be admin or normal"})
		return
	}

	passwordHash, err := hashSignupPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	user, team, err := a.insertUserWithTeam(req.Name, req.Email, passwordHash, teamName, utype, req.Verified)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if !user.IsVerified {
		if err := a.sendVerificationEmail(user, team.Name); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "user created but failed to send verification email: " + err.Error()})
			return
		}
	}

	c.JSON(http.StatusCreated, gin.H{
		"user": user,
		"team": team,
	})
}

func (a *CloudyApp) handleResetPassword(c *gin.Context) {
	var req resetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id, err := parseIDParam(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	user, err := a.getUserByID(id)
	if err != nil || user == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	hash, err := hashSignupPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err := a.updateUserPassword(user.ID, hash); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true, "user_id": user.ID})
}

func (a *CloudyApp) handleSetDisabled(c *gin.Context) {
	var req setDisabledRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id, err := parseIDParam(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	user, err := a.getUserByID(id)
	if err != nil || user == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	if err := a.setUserDisabled(user.ID, req.IsDisabled); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	unloaded := false
	if req.IsDisabled {
		inst, _, _ := a.getPrimaryInstanceForUser(user.ID)
		if inst != nil {
			unloaded = a.unloadSubApp(inst.Slug)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"ok":          true,
		"user_id":     user.ID,
		"is_disabled": req.IsDisabled,
		"unloaded":    unloaded,
	})
}

func (a *CloudyApp) redirectToApp(c *gin.Context) {
	name := strings.ToLower(c.Param("name"))
	claim := getClaim(c)
	if claim.UType != UTypeAdmin && claim.Slug != name {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}
	if !a.slugExists(name) {
		c.JSON(http.StatusNotFound, gin.H{"error": "instance not found"})
		return
	}

	host := fmt.Sprintf("%s.%s", name, normalizeDomain(a.config.Domain))
	c.Redirect(http.StatusFound, fmt.Sprintf("http://%s:%d/zz/pages", host, a.config.Port))
}

func (a *CloudyApp) loadApp(c *gin.Context) {
	name := strings.ToLower(c.Param("name"))
	if err := validateTenantSlug(name); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	claim := getClaim(c)
	if claim.UType != UTypeAdmin && claim.Slug != name {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	sub, err := a.ensureSubApp(name)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"tenant": name,
		"ready":  sub.IsReady(),
	})
}

func (a *CloudyApp) exportApp(c *gin.Context) {
	name := strings.ToLower(c.Param("name"))
	if err := validateTenantSlug(name); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	claim := getClaim(c)
	if claim.UType != UTypeAdmin && claim.Slug != name {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}
	if !a.slugExists(name) {
		c.JSON(http.StatusNotFound, gin.H{"error": "instance not found"})
		return
	}

	dbPath := filepath.Join(a.config.WorkingDir, "tenants", name, "app.db")
	if _, err := os.Stat(dbPath); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "database file not found"})
		return
	}

	a.mu.RLock()
	sub := a.subApps[name]
	a.mu.RUnlock()
	if sub != nil {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 20*time.Second)
		if err := sub.Checkpoint(ctx); err != nil {
			log.Printf("tenant %s: export checkpoint: %v", name, err)
		}
		cancel()
	}

	c.FileAttachment(dbPath, name+"-app-"+time.Now().Format("2006-01-02")+".db")
}

func (a *CloudyApp) unloadApp(c *gin.Context) {
	name := strings.ToLower(c.Param("name"))
	if err := validateTenantSlug(name); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if !a.slugExists(name) {
		c.JSON(http.StatusNotFound, gin.H{"error": "instance not found"})
		return
	}

	unloaded := a.unloadSubApp(name)
	c.JSON(http.StatusOK, gin.H{
		"tenant":   name,
		"unloaded": unloaded,
	})
}

func (a *CloudyApp) tenantRouteMW() gin.HandlerFunc {
	return func(c *gin.Context) {
		host := stripHostPort(c.Request.Host)
		tenant, isMain := a.parseTenantHost(host)
		if isMain || tenant == "" {
			c.Next()
			return
		}

		sub, err := a.ensureSubApp(tenant)
		if err != nil {
			status := http.StatusNotFound
			msg := "unknown instance"
			if strings.Contains(err.Error(), "disabled") {
				status = http.StatusForbidden
				msg = "instance is disabled"
			}
			c.JSON(status, gin.H{"error": msg})
			c.Abort()
			return
		}

		if sub.Engine == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "instance is not loaded"})
			c.Abort()
			return
		}
		sub.Engine.ServeHTTP(c.Writer, c.Request)
		c.Abort()
	}
}

func (a *CloudyApp) provisionTenant(user *User, slug string, adminPassword string) (*SubApp, error) {
	log.Println("provisionTenant/1", slug)

	remote, err := a.tenantRemote(slug)
	if err != nil {
		log.Println("provisionTenant/2", err)
		return nil, fmt.Errorf("provision tenant database: %w", err)
	}

	sub, err := NewSubApp(a.rootCtx, a.config, slug, true, a.mailer, remote)
	if err != nil {
		log.Println("provisionTenant/4", err)
		return nil, err
	}

	a.mu.Lock()
	if existing, ok := a.subApps[slug]; ok {
		a.mu.Unlock()
		if err := existing.WaitReady(30 * time.Second); err != nil {
			return nil, err
		}
		return existing, nil
	}

	a.subApps[slug] = sub
	a.mu.Unlock()

	if err := sub.Load(a.rootCtx, user.Fullname, adminPassword, user.Email); err != nil {
		a.mu.Lock()
		delete(a.subApps, slug)
		a.mu.Unlock()
		return nil, err
	}

	return sub, nil
}

func (a *CloudyApp) ensureSubApp(name string) (*SubApp, error) {
	inst, err := a.getPotatoInstanceBySlug(name)
	if err != nil {
		return nil, err
	}
	if inst == nil || inst.IsDeleted {
		return nil, fmt.Errorf("instance %q not found", name)
	}

	team, err := a.store.getTeamByID(inst.TeamID)
	if err != nil || team == nil || team.IsDeleted {
		return nil, fmt.Errorf("team for instance %q not active", name)
	}

	owner, err := a.getUserByID(team.OwnerID)
	if err != nil || owner == nil {
		return nil, fmt.Errorf("owner for instance %q not found", name)
	}
	if owner.IsDisabled {
		return nil, fmt.Errorf("owner for instance %q is disabled", name)
	}
	if !owner.IsVerified {
		return nil, fmt.Errorf("owner for instance %q is not verified", name)
	}

	a.mu.RLock()
	if sub, ok := a.subApps[name]; ok {
		a.mu.RUnlock()
		if err := sub.WaitReady(30 * time.Second); err != nil {
			return nil, err
		}
		return sub, nil
	}
	a.mu.RUnlock()

	dbPath := filepath.Join(a.config.WorkingDir, "tenants", name, "app.db")
	bootstrap := true
	if _, err := os.Stat(dbPath); err == nil {
		bootstrap = false
	}

	remote, err := a.tenantRemote(name)
	if err != nil {
		return nil, fmt.Errorf("resolve tenant database: %w", err)
	}

	sub, err := NewSubApp(a.rootCtx, a.config, name, bootstrap, a.mailer, remote)
	if err != nil {
		return nil, err
	}

	a.mu.Lock()
	if existing, ok := a.subApps[name]; ok {
		a.mu.Unlock()
		if err := existing.WaitReady(30 * time.Second); err != nil {
			return nil, err
		}
		return existing, nil
	}
	a.subApps[name] = sub
	a.mu.Unlock()

	if err := sub.Load(a.rootCtx, owner.Fullname, "", owner.Email); err != nil {
		a.mu.Lock()
		delete(a.subApps, name)
		a.mu.Unlock()
		return nil, err
	}

	return sub, nil
}

func (a *CloudyApp) unloadSubApp(name string) bool {
	a.mu.Lock()
	sub, ok := a.subApps[name]
	if ok {
		delete(a.subApps, name)
	}
	a.mu.Unlock()
	if !ok {
		return false
	}
	log.Printf("unloading tenant %s", name)
	sub.Unload()
	return true
}

func (a *CloudyApp) parseTenantHost(host string) (tenant string, isMain bool) {
	base := normalizeDomain(a.config.Domain)
	if host == base || host == "localhost" || host == "127.0.0.1" {
		return "", true
	}

	suffix := "." + base
	if strings.HasSuffix(host, suffix) {
		sub := strings.TrimSuffix(host, suffix)
		if sub == "" || strings.Contains(sub, ".") {
			return "", true
		}
		if reservedTenants[sub] {
			return "", true
		}
		return sub, false
	}

	return "", true
}

func validateTenantSlug(slug string) error {
	if slug == "" {
		return fmt.Errorf("slug is required")
	}
	if reservedTenants[slug] {
		return fmt.Errorf("slug %q is reserved", slug)
	}
	if !tenantSlugRe.MatchString(slug) {
		return fmt.Errorf("slug must be lowercase alphanumeric with hyphens, 2-63 chars")
	}
	return nil
}

func stripHostPort(host string) string {
	if h, _, err := net.SplitHostPort(host); err == nil {
		return h
	}
	return host
}

func parseIDParam(s string) (int64, error) {
	var id int64
	_, err := fmt.Sscanf(s, "%d", &id)
	return id, err
}
