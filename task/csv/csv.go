package csv

import (
	"encoding/csv"
	"fmt"
	engine "graph-engine"
	"io"
	"strings"
)

type Clause struct {
	column,
	value string
	index int
}

type CsvFilter struct {
	connectorId string
	clauses     []Clause
}

func NewCsvFilter(connectorId string, clauses ...Clause) *CsvFilter {
	return &CsvFilter{
		clauses:     clauses,
		connectorId: connectorId,
	}
}

func (cf *CsvFilter) Execute(gCtx *engine.GraphContext) error {
	conn, err := gCtx.GetConnector(cf.connectorId)
	if err != nil {
		return err
	}

	mockCsvConn, ok := conn.(*CSVConnector)
	if !ok {
		return fmt.Errorf("unable casting connector")
	}
	mockCsvConn.Filter(cf.clauses...)
	return nil
}

type MockCsvReader struct {
	b,
	connId string
}

func NewMockCsvReader(connId, content string) *MockCsvReader {
	return &MockCsvReader{
		connId: connId,
		b:      content,
	}
}

func (cr *MockCsvReader) Execute(gCtx *engine.GraphContext) error {
	conn, err := gCtx.GetConnector(cr.connId)
	if err != nil {
		return err
	}

	mockCsvConn, ok := conn.(*CSVConnector)
	if !ok {
		return fmt.Errorf("unable to casting connector")
	}

	mockCsvConn.LoadData(cr.b)
	return nil
}

type CSVConnector struct {
	column []string
	row    [][]string
}

func (c *CSVConnector) LoadData(data string) error {
	r := csv.NewReader(strings.NewReader(data))

	var (
		err error
	)

	c.column, err = r.Read()
	if err != nil {
		return err
	}

	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}

		if err != nil {
			return err
		}

		c.row = append(c.row, record)
	}

	return nil
}

func (c *CSVConnector) Filter(clauses ...Clause) {
	for i := range clauses {
		for j, column := range c.column {
			if column == clauses[i].column {
				clauses[i].index = j
				break
			}
		}
	}

	tempResult := make([][]string, 0)

	for _, r := range c.row {
		flag := true
		for _, c := range clauses {
			if r[c.index] != c.value {
				flag = false
				break
			}
		}

		if flag {
			tempResult = append(tempResult, r)
		}
	}

	c.row = tempResult
}

func (c *CSVConnector) GetRows() [][]string {
	return c.row
}

func (c *CSVConnector) String() string {
	sb := strings.Builder{}
	sb.WriteString(strings.Join(c.column, "|") + "\n")

	for _, r := range c.row {
		sb.WriteString(strings.Join(r, "|") + "\n")
	}

	return sb.String()
}
