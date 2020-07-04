package cmd

import (
	"fmt"
	"log"
	"os"
	"path"
	"strings"

	"github.com/jiharal/s1gu/utils"
	"github.com/spf13/cobra"
)

var modelORMCommand = &cobra.Command{
	Use:   "model-orm",
	Short: "Create model orm",
	Args:  cobra.MinimumNArgs(1),
	Run:   createModelORM,
}

func createModelORM(cmd *cobra.Command, args []string) {
	output := cmd.OutOrStderr()
	getEnv, _ := os.Getwd()

	pathModel := path.Join(getEnv, "model")

	if utils.IsExist(pathModel) {
		log.Print("Do you want to add it? [Yes|No] ")
		if !utils.AskForConfirmation() {
			os.Exit(2)
		}
	}

	filename := "model." + strings.ToLower(args[0]) + ".go"
	replacer := strings.NewReplacer(
		"{{.ModelName}}", strings.Title(args[0]),
		"{{.ModelNameLower}}", strings.ToLower(args[0]),
	)

	var fileContent = `
		package model

		import (
			"context"
			"fmt"
			"time"

			"github.com/go-pg/pg/v10"
			uuid "github.com/satori/go.uuid"
		)

		type (
			// {{.ModelName}} struct model
			{{.ModelName}}Model struct {
				ID          uuid.UUID ` + `pg:"type:uuid,default:gen_random_uuid(),pk"` + `
				Name      	string    ` + `pg:"name"` + `
				IsActive    bool    	` + `pg:"is_active"` + `
				CreatedAt   time.Time ` + `pg:"created_at,default:now()"` + `
				UpdatedAt   time.Time ` + `pg:"updated_at"` + `
				CreatedBy   uuid.UUID ` + `pg:"type:uuid"` + `
				UpdatedBy   uuid.UUID ` + `pg:"type:uuid"` + `
			}
			// {{.ModelName}}ModelResponse is a struct response.
			{{.ModelName}}ModelResponse struct {
				ID          uuid.UUID ` + `json:"id"` + `
				Name 				string		` + `json:"name"` + `
				IsActive		bool			` + `json:"is_active"` + `
				CreatedAt   time.Time ` + `json:"created_at"` + `
				UpdatedAt   time.Time ` + `json:"updated_at"` + `
				CreatedBy   uuid.UUID ` + `json:"created_by"` + `
				UpdatedBy   uuid.UUID ` + `json:"updated_by"` + `
			}
		)

		// Response Convert {{.ModelNameLower}} model into json-friendly formatted response struct (without null data type).
		func ({{.ModelNameLower}} *{{.ModelName}}Model) Response() {{.ModelName}}ModelResponse {
			return {{.ModelName}}ModelResponse{
				ID:          {{.ModelNameLower}}.ID,
				Name:        {{.ModelNameLower}}.Name,
				IsActive:    {{.ModelNameLower}}.IsActive,
				CreatedAt:   {{.ModelNameLower}}.CreatedAt,
				UpdatedAt:   {{.ModelNameLower}}.UpdatedAt,
				CreatedBy:   {{.ModelNameLower}}.CreatedBy,
				UpdatedBy:   {{.ModelNameLower}}.UpdatedBy,
			}
		}

		/ GetAll{{.ModelName}} is a ...
		func GetAll{{.ModelName}}(ctx context.Context, db *pg.DB, filter FilterOption) ([]{{.ModelName}}Model, error) {
			var {{.ModelName}}s []{{.ModelName}}Model
			if filter.Dir == "" || filter.Dir != "ASC" {
				filter.Dir = "DESC"
			} else {
				filter.Dir = "ASC"
			}
			err := db.Model(&{{.ModelName}}s).
				Where("name = CASE WHEN ? <> '' THEN ? ELSE name END", filter.Search, filter.Search).
				WhereOr("id = CASE WHEN ? <> '' THEN ? ELSE id END", uuid.FromStringOrNil(filter.Search), uuid.FromStringOrNil(filter.Search)).
				Order(fmt.Sprintf("created_at %s", filter.Dir)).
				Limit(filter.Limit).
				Offset(filter.Offset).
				Select()

			if err != nil {
				return nil, err
			}
			return {{.ModelName}}s, nil
		}

		// GetOne{{.ModelName}} is used to get one DB
		func GetOne{{.ModelName}}(ctx context.Context, db *pg.DB, param string) ({{.ModelName}}, error) {
			var {{.ModelName}} {{.ModelName}}
			err := db.Model(&{{.ModelName}}).
				Where("name = ?", param).
				WhereOr("id = ?", uuid.FromStringOrNil(param)).
				Select()

			if err != nil {
				return {{.ModelName}}, err
			}
			return {{.ModelName}}, nil
		}

		// Insert is used to ...
		func (data {{.ModelName}}) Insert(ctx context.Context, db *pg.DB) ({{.ModelName}}, error) {
			err := db.Insert(&data)
			if err != nil {
				return data, err
			}
			return data, nil
		}

		// Update is used to ...
		func ({{.ModelNameLower}} *{{.ModelName}}) Update(ctx context.Context, db *pg.DB) error {
			{{.ModelNameLower}}.UpdatedAt = time.Now()
			{{.ModelNameLower}} := db.Model({{.ModelNameLower}}).
				Set("name = ?", {{.ModelNameLower}}.Name).
				Set("updated_at = ?", {{.ModelNameLower}}.UpdatedAt)
			_, err := {{.ModelNameLower}}.Returning("*").
				Where("id = ?", {{.ModelNameLower}}.ID).
				Update()
			if err != nil {
				return err
			}
			return nil
		}

		// {{.ModelName}}Delete is used to delete by id
		func {{.ModelName}}Delete(ctx context.Context, db *pg.DB, id string) (*{{.ModelName}}, error) {
			var data {{.ModelName}}
			_, err := db.Model(&data).Where("id = ?", id).Delete()

			if err != nil {
				return nil, err
			}
			return data, nil
		}
	`
	fmt.Fprintf(output, "\t%s%screate%s\t %s%s\n", "\x1b[32m", "\x1b[1m", "\x1b[21m", path.Join(pathModel, filename), "\x1b[0m")
	utils.WriteToFile(path.Join(pathModel, filename), replacer.Replace(fileContent))

}