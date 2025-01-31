package clause

import (
	"fmt"
	"io"

	"github.com/stephenafamo/bob"
)

type CTE struct {
	Query        bob.Query // SQL standard says only select, postgres allows insert/update/delete
	Name         string
	Columns      []string
	Materialized *bool
	Search       CTESearch
	Cycle        CTECycle
}

func (c CTE) WriteSQL(w io.Writer, d bob.Dialect, start int) ([]any, error) {
	w.Write([]byte(c.Name))
	w.Write([]byte(" AS "))

	switch {
	case c.Materialized == nil:
		// do nothing
		break
	case *c.Materialized:
		w.Write([]byte("MATERIALIZED "))
	case !*c.Materialized:
		w.Write([]byte("NOT MATERIALIZED "))
	}

	args, err := bob.ExpressIf(w, d, start, c.Query, true, "(", ")")
	if err != nil {
		return nil, err
	}

	searchArgs, err := bob.ExpressIf(w, d, start+len(args), c.Search,
		len(c.Search.Columns) > 0, "\n", "")
	if err != nil {
		return nil, err
	}
	args = append(args, searchArgs...)

	cycleArgs, err := bob.ExpressIf(w, d, start+len(args), c.Cycle,
		len(c.Cycle.Columns) > 0, "\n", "")
	if err != nil {
		return nil, err
	}
	args = append(args, cycleArgs...)

	return args, nil
}

const (
	SearchBreadth = "BREADTH"
	SearchDepth   = "DEPTH"
)

type CTESearch struct {
	Order   string
	Columns []string
	Set     string
}

func (c CTESearch) WriteSQL(w io.Writer, d bob.Dialect, start int) ([]any, error) {
	// [ SEARCH { BREADTH | DEPTH } FIRST BY column_name [, ...] SET search_seq_col_name ]
	fmt.Fprintf(w, "SEARCH %s FIRST BY ", c.Order)

	args, err := bob.ExpressSlice(w, d, start, c.Columns, "", ", ", "")
	if err != nil {
		return nil, err
	}

	fmt.Fprintf(w, " SET %s", c.Set)

	return args, nil
}

type CTECycle struct {
	Columns    []string
	Set        string
	Using      string
	SetVal     any
	DefaultVal any
}

func (c CTECycle) WriteSQL(w io.Writer, d bob.Dialect, start int) ([]any, error) {
	//[ CYCLE column_name [, ...] SET cycle_mark_col_name [ TO cycle_mark_value DEFAULT cycle_mark_default ] USING cycle_path_col_name ]
	w.Write([]byte("CYCLE "))

	args, err := bob.ExpressSlice(w, d, start, c.Columns, "", ", ", "")
	if err != nil {
		return nil, err
	}

	fmt.Fprintf(w, " SET %s", c.Set)

	markArgs, err := bob.ExpressIf(w, d, start+len(args), c.SetVal,
		c.SetVal != nil, " TO ", "")
	if err != nil {
		return nil, err
	}
	args = append(args, markArgs...)

	defaultArgs, err := bob.ExpressIf(w, d, start+len(args), c.DefaultVal,
		c.DefaultVal != nil, " DEFAULT ", "")
	if err != nil {
		return nil, err
	}
	args = append(args, defaultArgs...)

	fmt.Fprintf(w, " USING %s", c.Using)

	return args, nil
}
