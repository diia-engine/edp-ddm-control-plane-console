package keycloak

import "context"

type ServiceInterface interface {
	GetUsers(ctx context.Context) ([]KeycloakRealmUser, error)
	CreateUser(ctx context.Context, user *KeycloakRealmUser) error
	UpdateUser(ctx context.Context, user *KeycloakRealmUser) error
	DeleteUser(ctx context.Context, user *KeycloakRealmUser) error
	GetUsersByRealm(ctx context.Context, realmName string) ([]KeycloakRealmUser, error)
}
