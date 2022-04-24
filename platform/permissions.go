package platform

// List of all registered permissions from all the modules
var AllPermissions = NewPermissions()

type Permission string

type Permissions []Permission

func NewPermissions() *Permissions {
	p := make(Permissions, 0)
	return &p
}

func (t *Permissions) Grant(list ...Permission) {
	if t == nil {
		return
	}
	for _, p := range list {
		*t = append(*t, p)
	}
}

func (t *Permissions) GrantAll() {
	t.Grant(*AllPermissions...)
}

func (t *Permissions) Revoke(per ...Permission) {
	if t == nil {
		return
	}
	for _, p := range per {
		for idx, name := range *t {
			if name == p {
				if idx == len(*t)-1 {
					*t = (*t)[:idx]
				} else {
					*t = append((*t)[:idx], (*t)[idx+1:]...)
				}
			}
		}
	}
}

func (t *Permissions) RevokeAll() {
	*t = (*t)[:0]
}

func (t *Permissions) Can(per Permission) bool {
	if t == nil {
		return false
	}
	for _, perm := range *t {
		if perm == per {
			return true
		}
	}
	return false
}

func (t *Permissions) CanAny(list ...Permission) bool {
	if t == nil {
		return false
	}
	for _, perm := range list {
		if t.Can(perm) {
			return true
		}
	}
	return false
}

func (t *Permissions) CanAll(list ...Permission) bool {
	if t == nil {
		return false
	}
	for _, perm := range list {
		if !t.Can(perm) {
			return false
		}
	}
	return true
}

func RegisterPermissions(name ...Permission) {
	AllPermissions.Grant(name...)
}
