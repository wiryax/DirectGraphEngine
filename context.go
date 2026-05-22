package engine

type GraphContext struct {
	rState    *RuntimeState
	logger    GraphLogger
	connector *Connector
}

func NewGraphContext(gLog GraphLogger, rState *RuntimeState) *GraphContext {
	return &GraphContext{
		logger: gLog,
		rState: rState,
	}
}

func NewGraphContextWithConnector(gLog GraphLogger, rState *RuntimeState, connector *Connector) *GraphContext {
	return &GraphContext{
		logger:    gLog,
		rState:    rState,
		connector: connector,
	}
}

func (gCtx *GraphContext) GetVariable(key string) (string, error) {
	return gCtx.rState.GetVariable(key)
}

func (gCtx *GraphContext) SetVariable(key, value string) {
	gCtx.rState.SetVariable(key, value)
}

func (gCtx *GraphContext) GetConnector(key string) (any, error) {
	return gCtx.connector.GetConnector(key)
}

func (gCtx *GraphContext) SetConnector(key string, conn any) {
	gCtx.connector.SetConnector(key, conn)
}

func (gCtx *GraphContext) Log(et EvenType, logLv LogLevel, msg, vId, gId string) {
	gCtx.logger.FlushLog(et, logLv, msg, vId, gId)
}
