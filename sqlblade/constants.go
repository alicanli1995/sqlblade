package sqlblade

const (
	dialectPostgres = "postgres"

	// Buffer sizes for SQL building
	sqlBuilderBufferSize  = 512
	argsInitialCapacity   = 8
	resultInitialCapacity = 10
	updateBufferSize      = 256

	insertBufferSize      = 1024
	selectBufferSize      = 768
	batchInsertBufferSize = 2048
)
