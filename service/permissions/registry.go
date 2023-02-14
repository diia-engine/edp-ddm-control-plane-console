package permissions

import (
	"ddm-admin-console/router"
	"ddm-admin-console/service/codebase"
	"ddm-admin-console/service/k8s"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type RegistryPermission struct {
	CanGet    bool
	CanUpdate bool
	CanDelete bool
	Expiry    time.Time
}

type Registry struct {
	//token - []permission
	perms           map[string]map[string]RegistryPermission
	permsLock       sync.RWMutex
	codebaseService codebase.ServiceInterface
	k8sService      k8s.ServiceInterface
}

func Make(cbService codebase.ServiceInterface, k8sService k8s.ServiceInterface) *Registry {
	r := Registry{
		perms:           make(map[string]map[string]RegistryPermission),
		k8sService:      k8sService,
		codebaseService: cbService,
	}

	//go r.expiryTicker()

	return &r
}

func (r *Registry) expiryTicker() {
	tk := time.NewTicker(time.Second)

	for range tk.C {
		r.CheckExpiry()
	}
}

func (r *Registry) CheckExpiry() {
	r.permsLock.RLock()

	var tokensToRemove []string
	for tok, regPerms := range r.perms {
		for _, perms := range regPerms {
			if time.Now().Unix() > perms.Expiry.Unix() {
				tokensToRemove = append(tokensToRemove, tok)
			}

			break
		}
	}

	r.permsLock.RUnlock()

	r.permsLock.Lock()
	for _, t := range tokensToRemove {
		delete(r.perms, t)
	}
	r.permsLock.Unlock()
}

func (r *Registry) SetPermission(token string, registryName string, permission RegistryPermission) {
	if time.Now().Unix() > permission.Expiry.Unix() {
		return
	}

	r.permsLock.Lock()
	tokenPerms, ok := r.perms[token]
	if !ok {
		tokenPerms = make(map[string]RegistryPermission)
	}

	tokenPerms[registryName] = permission

	r.perms[token] = tokenPerms
	r.permsLock.Unlock()
}

func (r *Registry) GetPermission(token, registryName string) (*RegistryPermission, error) {
	r.permsLock.RLock()

	regs, ok := r.perms[token]
	if !ok {
		r.permsLock.RUnlock()
		return nil, errors.New("no permission")
	}

	perm, ok := regs[registryName]
	if !ok {
		r.permsLock.RUnlock()
		return nil, errors.New("no permission")
	}

	if time.Now().Unix() > perm.Expiry.Unix() {
		r.permsLock.RUnlock()
		r.DeleteToken(token)
		return nil, errors.New("no permission")
	}

	r.permsLock.RUnlock()
	return &perm, nil
}

func (r *Registry) DeleteTokenContext(ctx *gin.Context) error {
	tok, err := router.ExtractToken(ctx)
	if err != nil {
		return fmt.Errorf("no token: %w", err)
	}

	r.DeleteToken(tok.AccessToken)

	return nil
}

func (r *Registry) DeleteToken(tok string) {
	r.permsLock.Lock()
	defer r.permsLock.Unlock()

	delete(r.perms, tok)
}

func (r *Registry) DeleteRegistry(name string) {
	r.permsLock.Lock()
	defer r.permsLock.Unlock()

	for token, regPerms := range r.perms {
		_, ok := regPerms[name]
		if ok {
			delete(r.perms[token], name)
			return
		}
	}
}

func (r *Registry) FilterCodebases(ginContext *gin.Context, cbs []codebase.Codebase, k8sService k8s.ServiceInterface) ([]codebase.WithPermissions, error) {
	tok, err := router.ExtractToken(ginContext)
	if err != nil {
		return nil, fmt.Errorf("no token: %w", err)
	}

	withPerms := make([]codebase.WithPermissions, 0, len(cbs))
	for i, cb := range cbs {
		perm, err := r.GetPermission(tok.AccessToken, cb.Name)
		if err == nil {
			if !perm.CanGet {
				continue
			}
		} else {
			canGet, canUpdate, canDelete, err := codebase.CheckCodebasePermission(cb.Name, k8sService)
			if err != nil {
				return nil, fmt.Errorf("unable to check perms: %w", err)
			}

			perm = &RegistryPermission{CanUpdate: canUpdate, CanDelete: canDelete, CanGet: true, Expiry: tok.Expiry}
			r.SetPermission(tok.AccessToken, cb.Name, *perm)

			if !canGet {
				continue
			}
		}

		withPerms = append(withPerms, codebase.WithPermissions{
			Codebase:  &cbs[i],
			CanUpdate: perm.CanUpdate,
			CanDelete: perm.CanDelete,
		})
	}

	return withPerms, nil
}

func (r *Registry) LoadUserRegistries(ctx *gin.Context) error {
	cbs, err := r.codebaseService.GetAllByType(codebase.RegistryCodebaseType)
	if err != nil {
		return fmt.Errorf("unable to load codebases: %w", err)
	}

	k8sService, err := r.k8sService.ServiceForContext(ctx)
	if err != nil {
		return fmt.Errorf("unable to get k8s service for context: %w", err)
	}

	tokenData, err := router.ExtractToken(ctx)
	if err != nil {
		return fmt.Errorf("unable to extract token from context, err: %w", err)
	}

	for _, cb := range cbs {
		canGet, canUpdate, canDelete, err := codebase.CheckCodebasePermission(cb.Name, k8sService)
		if err != nil {
			return fmt.Errorf("unable to check codebase permissions: %w", err)
		}

		r.SetPermission(tokenData.AccessToken, cb.Name, RegistryPermission{
			CanUpdate: canUpdate,
			CanGet:    canGet,
			CanDelete: canDelete,
			Expiry:    tokenData.Expiry,
		})

	}

	return nil
}
