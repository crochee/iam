package storage

type Base struct {
	ID        uint64    `json:"id,string" gorm:"primary_key:id"`
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at;not null;comment:创建时间"`
	UpdatedAt time.Time `json:"updated_at" gorm:"column:updated_at;not null;comment:更新时间"`
	DeletedAt DeletedAt `json:"-" gorm:"column:deleted_at;index;comment:删除时间"`
}

func (b *Base) BeforeCreate(db *gorm.DB) error {
	if b.ID == 0 {
		snowID, err := idx.NextID()
		if err != nil {
			return err
		}
		b.ID = snowID
	}
	return nil
}
