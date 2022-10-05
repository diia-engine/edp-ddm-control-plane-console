package registry

func (a *App) createRoutes() {
	a.router.GET("/admin/registry/create", a.createRegistryGet)
	a.router.POST("/admin/registry/create", a.createRegistryPost)

	a.router.GET("/admin/registry/edit/:name", a.editRegistryGet)
	a.router.POST("/admin/registry/edit/:name", a.editRegistryPost)

	a.router.GET("/admin/registry/overview", a.listRegistry)
	a.router.POST("/admin/registry/overview", a.deleteRegistry)

	a.router.GET("/admin/registry/check/:name", a.registryNameAvailable)
	a.router.GET("/admin/registry/view/:name", a.viewRegistry)
	a.router.POST("/admin/registry/update/:name", a.registryUpdate)

	a.router.POST("/admin/registry/external-reg-add/:name", a.addExternalReg)
	a.router.POST("/admin/registry/external-reg-remove/:name", a.removeExternalReg)
	a.router.POST("/admin/registry/external-reg-disable/:name", a.disableExternalReg)

	a.router.GET("/admin/registry/:name/change/:change", a.viewChange)
}
