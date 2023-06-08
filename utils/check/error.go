package check

func Error(err error, isFatal bool, isVerbose bool) {
	if err != nil {
		if isVerbose {
			println(err.Error())
		}
		if isFatal {
			panic(err)
		}
	}
}
