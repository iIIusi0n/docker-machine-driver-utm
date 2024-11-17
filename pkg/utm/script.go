package utm

import "github.com/andybrewer/mack"

const UtmAppName = "UTM"

func runUtmScript(script ...string) (string, error) {
	return mack.Tell(UtmAppName, script...)
}
