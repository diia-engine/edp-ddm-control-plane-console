package cluster

func (a *App) createRoutes() {
	a.router.GET("/admin/cluster/management", a.view)

	a.router.GET("/admin/cluster/edit", a.editGet)
	a.router.POST("/admin/cluster/edit", a.editPost)
	a.router.POST("/admin/cluster/upgrade", a.clusterUpdate)
	a.router.POST("/admin/cluster/admins", a.updateAdminsView)
	a.router.POST("/admin/cluster/key", a.updateKeyView)
	a.router.POST("/admin/cluster/cidr", a.updateCIDRView)
	a.router.POST("/admin/cluster/backup-schedule", a.backupSchedule)
	a.router.POST("/admin/cluster/upload-pem-dns", a.uploadPEMDNS)
	a.router.POST("/admin/cluster/add-keycloak-dns", a.keycloakDNS)
	a.router.GET("/admin/cluster/check-keycloak-hostname/:hostname", a.checkKeycloakHostnameUsed)
}
