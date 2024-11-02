package studyengine

import(
	"time"
)

// Now function control the current time used by the expressions. 
var Now func() time.Time = time.Now
