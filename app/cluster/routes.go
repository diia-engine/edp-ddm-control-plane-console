package cluster

func (a *App) createRoutes() {
	a.router.GET("/admin/cluster/management", a.view)

	a.router.GET("/admin/cluster/edit", a.editGet)
	a.router.POST("/admin/cluster/edit", a.editPost)
	a.router.POST("/admin/cluster/upgrade", a.clusterUpdate)
	a.router.POST("/admin/cluster/admins", a.updateAdminsView)
	a.router.POST("/admin/cluster/key", a.updateKeyView)
}
