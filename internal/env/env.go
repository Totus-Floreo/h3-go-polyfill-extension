package env

type Env struct {
	targetTable  string `env:"TARGET_TABLE,required"`
	outlineTable string `env:"OUTLINE_TABLE,required"`
	dbConnStr    string `env:"DB_CONN_STR,required"`
}

func NewEnv(targetTable, outlineTable, dbConnStr string) Env {
	return Env{
		targetTable:  targetTable,
		outlineTable: outlineTable,
		dbConnStr:    dbConnStr,
	}
}

func (e *Env) TargetTable() string {
	return e.targetTable
}

func (e *Env) OutlineTable() string {
	return e.outlineTable
}
func (e *Env) DbConnStr() string {
	return e.dbConnStr
}
