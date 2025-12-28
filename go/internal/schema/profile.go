// pkg/vax/schemas/profile.go
package schemas

import "vax/internal/schema"

type UpdateProfileDTO struct {
    DisplayName string  `json:"displayName" validate:"required,max=50" description:"User's display name"`
    Bio         *string `json:"bio,omitempty" validate:"omitempty,max=200" description:"User bio"`
    Avatar      *string `json:"avatar,omitempty" validate:"omitempty,url" description:"Avatar URL"`
}

func GetUpdateProfileSchema() map[string]interface{} {
    return schema.Generate[UpdateProfileDTO]()
}
