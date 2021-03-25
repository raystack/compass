package lineage

import (
	"fmt"

	"github.com/PaesslerAG/jsonpath"
	"github.com/odpf/columbus/lib/set"
	"github.com/odpf/columbus/models"
)

type lineageProcessor struct {
	descriptors []models.LineageDescriptor
}

func (lp lineageProcessor) LineageOf(record models.Record) (upstream, downstream set.StringSet, err error) {
	upstream, downstream = set.NewStringSet(), set.NewStringSet()
	for _, desc := range lp.descriptors {
		var values []string
		values, err = lp.executeQuery(record, desc.Query)
		if err != nil {
			return
		}
		values = lp.addTypePrefix(values, desc.Type)

		switch desc.Dir {
		case models.DataflowDirDownstream:
			lp.updateSetWithValues(downstream, values)
		case models.DataflowDirUpstream:
			lp.updateSetWithValues(upstream, values)
		default:
			err = fmt.Errorf("unsupported dataflow dir = %q", desc.Dir)
			return
		}
	}
	return
}

func (lp lineageProcessor) executeQuery(data models.Record, query string) (values []string, err error) {

	defer func() {
		if err != nil {
			err = fmt.Errorf("executeQuery: %v", err)
		}
	}()

	result, err := jsonpath.Get(query, data)
	if err != nil {
		return
	}
	values, err = lp.parseQueryResultAsStringSlice(result, query)
	if err != nil {
		return
	}
	return values, nil
}

func (lp lineageProcessor) parseQueryResultAsStringSlice(result interface{}, query string) (values []string, err error) {

	defer func() {
		if err != nil {
			err = fmt.Errorf("parseQueryResultAsStringSlice: %v", err)
		}
	}()

	switch raw := result.(type) {
	case string:
		values = []string{raw}
	case []string:
		values = raw
	case []interface{}:
		// in case the query evaluates to a list of ambigious type,
		// check if the values are string. Sometimes jsonpath can produce
		// these lists if you query embedded objects.
		for _, v := range raw {
			val, isString := v.(string)
			if !isString {
				err = fmt.Errorf("expected value to be string, was %T", v)
				return
			}
			values = append(values, val)
		}
	default:
		template := "query must evaluate to a string or an array of string. expression=%s, value type=%T"
		err = fmt.Errorf(template, query, raw)
		return
	}
	return
}

func (lp lineageProcessor) updateSetWithValues(ss set.StringSet, values []string) {
	for _, value := range values {
		ss.Add(value)
	}
}

func (lp lineageProcessor) addTypePrefix(values []string, typ string) (rv []string) {
	for _, value := range values {
		rv = append(rv, fmt.Sprintf("%s/%s", typ, value))
	}
	return
}

func newLineageProcessor(cfg []models.LineageDescriptor) lineageProcessor {
	return lineageProcessor{
		descriptors: cfg,
	}
}
