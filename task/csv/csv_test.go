package csv

import (
	"fmt"
	engine "graph-engine"
	"os"
	"reflect"
	"testing"
)

func TestCsvFilterWorkflow(t *testing.T) {
	logger := engine.NewLogger(os.Stdout)
	mockCsvConnKey := "mockCsvConn"
	rState := engine.NewRuntimeState(make(map[string]string))
	connector := engine.NewConnector(make(map[string]any))
	gCtx := engine.NewGraphContextWithConnector(logger, rState, connector)
	gCtx.SetConnector(mockCsvConnKey, &CSVConnector{column: make([]string, 0), row: make([][]string, 0)})

	graph := engine.NewGraph("TestCsvWorkflow")

	csvReader := graph.Add("Csv Reader", NewMockCsvReader(mockCsvConnKey, "first_name,middle_name,last_name\nwirya,muhammad,nugraha\nnugraha,muhammad,wirya"))
	csvFilter := graph.Add("Csv Filter", NewCsvFilter(mockCsvConnKey,
		Clause{
			column: "first_name",
			value:  "wirya",
		}, Clause{
			column: "middle_name",
			value:  "muhammad",
		}, Clause{
			column: "last_name",
			value:  "nugraha",
		}),
	)

	graph.Connect(csvReader, csvFilter, engine.Success, engine.ExpAnd, nil)

	graph.RunWithContext(gCtx)

	expectedResult := [][]string{
		{"wirya", "muhammad", "nugraha"},
	}

	conn, err := gCtx.GetConnector(mockCsvConnKey)
	if err != nil {
		t.Fatalf("unexpected error occur, %v", err)
	}

	mockCsvConn, ok := conn.(*CSVConnector)
	if !ok {
		t.Fatalf("unable casting connector")
	}

	if !reflect.DeepEqual(fmt.Sprintf("%v", expectedResult), fmt.Sprintf("%v", mockCsvConn.GetRows())) {
		t.Errorf("unexpected result. want %v got %v", expectedResult, mockCsvConn.GetRows())
	}
}
