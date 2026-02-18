package domain

type ChangeTracker struct {
	dirtyFields map[string]bool
}

func NewChangeTracker() *ChangeTracker {
	return &ChangeTracker{
		dirtyFields: make(map[string]bool),
	}
}

func (ct *ChangeTracker) MarkDirty(field string) {
	ct.dirtyFields[field] = true
}

func (ct *ChangeTracker) Dirty(field string) bool {
	return ct.dirtyFields[field]
}

func (ct *ChangeTracker) HasChanges() bool {
	return len(ct.dirtyFields) > 0
}

func (ct *ChangeTracker) Reset() {
	ct.dirtyFields = make(map[string]bool)
}

func (ct *ChangeTracker) GetDirtyFields() []string {
	fields := make([]string, 0, len(ct.dirtyFields))
	for field := range ct.dirtyFields {
		fields = append(fields, field)
	}
	return fields
}