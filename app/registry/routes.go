package registry

func (a *App) createRoutes() {
	a.router.GET("/admin/registry/create", a.createRegistryGet)
	a.router.POST("/admin/registry/create", a.createRegistryPost)

	a.router.GET("/admin/registry/edit/:name", a.editRegistryGet)
	a.router.POST("/admin/registry/edit/:name", a.editRegistryPost)

	a.router.GET("/admin/registry/overview", a.listRegistry)
	a.router.POST("/admin/registry/overview", a.deleteRegistry)

	a.router.POST("/admin/registry/check-pem", a.validatePEMFile)
	a.router.GET("/admin/registry/check/:name", a.registryNameAvailable)
	a.router.GET("/admin/registry/view/:name", a.viewRegistry)
	a.router.POST("/admin/registry/update/:name", a.registryUpdate)
	a.router.GET("/admin/registry/update/:name", a.registryUpdateView)
	a.router.POST("/admin/registry/trembita-client/:name", a.setTrembitaClientRegistryData)
	a.router.POST("/admin/registry/trembita-client-create/:name", a.createTrembitaClientRegistry)
	a.router.GET("/admin/registry/trembita-client-check/:name", a.checkTrembitaClientExists)
	a.router.GET("/admin/registry/trembita-client-delete/:name", a.deleteTrembitaClient)
	a.router.POST("/admin/registry/external-system/:name", a.setExternalSystemRegistryData)
	a.router.POST("/admin/registry/external-system-create/:name", a.createExternalSystemRegistry)
	a.router.GET("/admin/registry/external-system-check/:name", a.checkExternalSystemExists)
	a.router.GET("/admin/registry/external-system-delete/:name", a.deleteExternalSystem)

	a.router.POST("/admin/registry/external-reg-add/:name", a.addExternalReg)
	a.router.POST("/admin/registry/external-reg-remove/:name", a.removeExternalReg)
	a.router.POST("/admin/registry/external-reg-disable/:name", a.disableExternalReg)

	a.router.GET("/admin/change/:change", a.viewChange)
	a.router.GET("/admin/submit-change/:change", a.submitChange)
	a.router.GET("/admin/abandon-change/:change", a.abandonChange)

	a.router.GET("/admin/registry/preload-resources", a.preloadTemplateResources)
	a.router.GET("/admin/registry/get-basic-username/:name", a.getBasicUsername)
}
