package flags

import (
	"flag"
	"strings"
)

func Setup() (bool, bool, uint, []string, []string, string) {
	var debug bool
	var dryrun bool
	var skipNewerThanDays uint
	var tag string
	var regions []string
	var nodegroups []string

	flag.BoolVar(&debug, "debug", false, "set log level to debug (eg. '--debug=true')")
	flag.BoolVar(&dryrun, "dryrun", false, "set dryrun mode (eg. '--dryrun=true')")
	flag.UintVar(&skipNewerThanDays, "skip-newer-than-days", 0, "skip ami update if the latest available ami was published in less than provided number of days (eg. '--skip-newer-than-days=7')")
	flag.StringVar(&tag, "tag", "", "update amis only for nodegroups within this tag (eg. '--tag=env:production')")
	flag.Func("nodegroups", "update amis for (only specified here) nodegroups (eg. '--nodegroups=eu-west-1:cluster-1:ngMain,eu-west-2:clusterStage:nodegroupStage1')", func(s string) error {
		nodegroups = strings.Split(s, ",")

		return nil
	})
	flag.Func("regions", "update amis for all nodegroups from those regions only (eg. '--regions=eu-west-1,us-west-1')", func(s string) error {
		regions = strings.Split(s, ",")

		return nil
	})
	flag.Parse()

	return debug, dryrun, skipNewerThanDays, regions, nodegroups, tag
}
