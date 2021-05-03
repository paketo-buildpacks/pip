package pip_test

import (
	"testing"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestUnitPip(t *testing.T) {
	suite := spec.New("pip", spec.Report(report.Terminal{}))
	suite("Detect", testDetect)
	suite("Build", testBuild)
	suite("InstallProcess", testPipInstallProcess)
	suite("SiteProcess", testSiteProcess)
	suite.Run(t)
}
