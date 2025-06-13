package schema

import (
	"context"
	"errors"
	"regexp"
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"github.com/yc-alpha/admin/app/user_management/internal/data/ent/hook"
	"github.com/yc-alpha/admin/common/snowflake"
	"golang.org/x/crypto/bcrypt"
)

// SysUser holds the schema definition for the SysUser entity.
type SysUser struct {
	ent.Schema
}

// Fields of the SysUser.
func (SysUser) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").
			Immutable().
			Unique().
			DefaultFunc(func() int64 {
				return snowflake.Generate().Int64()
			}).
			Comment("Primary Key ID"),
		field.String("username").
			MaxLen(64).
			NotEmpty().
			Unique().
			Comment("Username of the user"),
		field.String("email").
			Match(regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)). // 基本 email 格式校验
			NotEmpty().
			Nillable().
			Unique().
			Comment("Email address of the user"),
		// E.164 国际手机号格式，例如 +8613812345678
		field.String("phone").
			Match(regexp.MustCompile(`^\+[1-9]\d{1,14}$`)).
			MaxLen(16).
			NotEmpty().
			Nillable().
			Unique().
			Comment("Phone number of the user"),
		field.String("password").
			NotEmpty().
			Nillable().
			Sensitive().
			Comment("Password of the user"),
		field.Enum("status").
			Values("active", "disabled", "pending").
			Default("pending").
			Comment("Status of the user"),
		field.String("full_name").
			Nillable().
			Comment("Full name of the user"),
		field.Enum("gender").
			Values("male", "female", "unknown").
			Default("unknown").
			Comment("User gender"),
		field.String("avatar").
			Nillable().
			Comment("Avatar URL of the user"),
		field.String("language").
			Default("en").
			Match(regexp.MustCompile(`^(en|zh|fr|es|de|ja|ko)$`)).
			Comment("Preferred language of the user"),
		field.String("timezone").
			Default("Asia/Shanghai").
			Validate(func(timezone string) error {
				// Validate that the timezone is a valid IANA timezone
				if _, err := time.LoadLocation(timezone); err != nil {
					return err
				}
				return nil
			}).
			Comment("Preferred timezone of the user"),
		field.Time("last_login").
			Nillable().
			Comment("Timestamp of the last login by the user"),
		field.String("last_ip").
			Nillable().
			Comment("IP address of the last login by the user"),
		field.Int64("created_by").
			Nillable().
			Comment("User who created this record"),
		field.Int64("updated_by").
			Nillable().
			Comment("User who last updated this record"),
		field.Time("created_at").
			Default(time.Now).
			Immutable().
			Comment("Creation timestamp of the user record"),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now).
			Comment("Last update timestamp of the user record"),
		field.Time("deleted_at").
			Optional().
			Nillable().
			Comment("Timestamp when the user was deleted, if applicable"),
	}
}

// Edges of the SysUser.
func (SysUser) Edges() []ent.Edge {
	return nil
}

// hashes the password before saving it to the database.
func (SysUser) Hooks() []ent.Hook {
	return []ent.Hook{
		hook.On(func(next ent.Mutator) ent.Mutator {
			return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
				if password, ok := m.Field("password"); ok && password != nil {
					if passStr, ok := password.(string); ok {
						// Hash the password before saving
						hashedPassword, err := bcrypt.GenerateFromPassword([]byte(passStr), bcrypt.DefaultCost)
						if err != nil {
							return nil, err
						}
						m.SetField("password", hashedPassword)
					} else {
						return nil, errors.New("password must be a string")
					}
				}
				return next.Mutate(ctx, m)
			})
		}, ent.OpCreate|ent.OpUpdate|ent.OpUpdateOne),
	}
}
