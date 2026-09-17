package origins

func Compare(public Interface, implementations map[string]Function, call CallCase) Comparison {
	consumer := ConsumerCheck(public, call)
	implementation, ok := implementations[call.Function]
	if !ok {
		oracle := invalid(call, "oracle.unknown_function", "implementation is absent")
		return Comparison{Case: call.ID, Agreement: false, Consumer: consumer, Oracle: oracle}
	}
	oracle := OracleCheck(public.Types, implementation, call)
	return Comparison{
		Case:      call.ID,
		Agreement: consumer.Valid == oracle.Valid && diagnosticCode(consumer) == diagnosticCode(oracle),
		Consumer:  consumer,
		Oracle:    oracle,
	}
}

func diagnosticCode(result CheckResult) string {
	if result.Diagnostic == nil {
		return ""
	}
	return result.Diagnostic.Code
}
