package vault

// Config defines Vault-related configuration.
type Config struct {
	// Address is the Vault server address.
	Address *string
	// Token is the authentication token for Vault.
	Token *string
}
