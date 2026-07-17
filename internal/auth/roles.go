package auth

// Role values stored on users.role. Kept as plain strings (matching the DB
// column type) rather than a Go enum, but centralized here so call sites
// don't repeat string literals.
const (
	RoleCustomer = "customer"
	RoleBarber   = "barber"
	RoleFounder  = "founder"
	RoleAdmin    = "admin"
)

// ValidRole reports whether role is one of the known role values.
func ValidRole(role string) bool {
	switch role {
	case RoleCustomer, RoleBarber, RoleFounder, RoleAdmin:
		return true
	default:
		return false
	}
}
