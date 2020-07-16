package db

import (
	"database/sql"
	"fmt"
	"os"
	"strings"
	"unicode"

	"github.com/spf13/cobra"

	_ "github.com/godror/godror"
)

var OracleCmd = &cobra.Command{
	Use:   "oracle",
	Short: "将Oracle表结构转换成Golang的结构",
	Long:  "将Oracle表结构转换成Golang的结构",
	RunE: func(cmd *cobra.Command, args []string) error {
		err := ReGenDir(OracleEngine.output)
		if err != nil {
			return err
		}

		err = OracleEngine.connect()
		if err != nil {
			return err
		}

		if len(OracleEngine.tableNames) == 0 {
			err := OracleEngine.Generates()
			if err != nil {
				return err
			}
		} else {
			for _, tabName := range OracleEngine.tableNames {
				err := OracleEngine.Generate(tabName)
				if err != nil {
					return err
				}
			}
		}

		return nil
	},
}

var OracleEngine *OrcEngine

func init() {
	OracleEngine = &OrcEngine{}

	OracleCmd.Flags().StringVarP(&OracleEngine.dns, "dns", "d", "", "连接Oracle数据库的字符串，如：user/passw@service_name")
	OracleCmd.Flags().StringVarP(&OracleEngine.pkg, "pkg", "p", "models", "生成文件的包名")
	OracleCmd.Flags().StringArrayVarP(&OracleEngine.tableNames, "tables", "t", []string{},
		`需要转换的表名（不区分大小写），如果不填，默认转换该用户下所有表
多个表时这样写：-t tableA -t tableB -t tableC	
		`)
	OracleCmd.Flags().StringVarP(&OracleEngine.output, "output", "o", "./models", "文件生成输出的目录")
}

type OrcEngine struct {
	db *sql.DB

	dns        string
	output     string
	pkg        string
	tableNames []string
}

func (o *OrcEngine) connect() error {
	var err error
	o.db, err = sql.Open("godror", o.dns)
	if err != nil {
		return err
	}

	err = o.db.Ping()
	if err != nil {
		return err
	}

	return nil
}

func (o *OrcEngine) Generates() error {
	rows, err := o.db.Query("SELECT TABLE_NAME FROM user_tab_comments")
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var tableName string
		err = rows.Scan(&tableName)
		if err != nil {
			fmt.Println(err.Error())
			return err
		}

		err = o.Generate(tableName)
		if err != nil {
			return err
		}
	}

	err = rows.Err()
	if err != nil {
		fmt.Println(err.Error())
		return err
	}

	return nil
}

func (o *OrcEngine) getTableInfo(tableName string) ([]*StructColumn, error) {
	rows, err := o.db.Query(
		`SELECT A.COLUMN_NAME, A.DATA_TYPE, B.COMMENTS
FROM USER_TAB_COLUMNS A
         INNER JOIN USER_COL_COMMENTS B
                    ON A.TABLE_NAME = B.TABLE_NAME AND A.TABLE_NAME = :1 AND A.COLUMN_NAME = B.COLUMN_NAME`, tableName)
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}
	defer rows.Close()

	list := make([]*StructColumn, 0)

	for rows.Next() {
		info := &StructColumn{}
		var comment sql.NullString
		err = rows.Scan(
			&info.Name,
			&info.DataType,
			&comment,
		)

		if err != nil {
			fmt.Println(err.Error())
			return nil, err
		}

		if comment.Valid {
			info.Comment = comment.String
		}

		info.Name = UnderscoreToUpperCamelCase(info.Name)
		info.Tag = fmt.Sprintf("`json:\"%s\"`", string(unicode.ToLower(rune(info.Name[0])))+info.Name[1:])
		info.DataType = o.ConvertType(info.DataType)

		list = append(list, info)
	}

	if err = rows.Err(); err != nil {
		fmt.Println(err.Error())
		return nil, err
	}

	return list, nil
}

func (o *OrcEngine) ConvertType(tp string) string {
	switch tp {
	case "VARCHAR2", "VARCHAR", "NVARCHAR2", "NVARCHAR", "CHAR", "NCHAR":
		return "string"
	case "NUMBER", "LONG", "TIMESTAMP", "ROWID", "UROWID":
		return "int64"
	case "FLOAT", "BINARY_FLOAT", "BINARY_DOUBLE":
		return "float64"
	case "DATE":
		return "time.Time"
	case "RAW", "LONG RAW":
		return "[]byte"
	case "CLOB", "NCLOB", "BLOB", "BFILE":
		return "interface{}"
	}

	return "interface{}"
}

func (o *OrcEngine) Generate(tableName string) error {
	tableName = strings.ToUpper(tableName)
	list, err := o.getTableInfo(tableName)
	if err != nil {
		return err
	}

	if len(list) == 0 {
		return fmt.Errorf("no data")
	}

	tpl := GetTemplate()

	tplDB := StructTemplateDB{
		TableName: tableName,
		Columns:   list,
		Package:   o.pkg,
	}

	file, err := os.OpenFile(fmt.Sprintf("%s/gen___%s.go",
		o.output, strings.ToLower(tableName)), os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	defer file.Close()

	err = tpl.Execute(file, tplDB)
	if err != nil {
		fmt.Println(err)
		return err
	}

	fmt.Printf("已生成[%s]\n", tableName)

	return nil
}
