package main

import (
	"github.com/rs/zerolog/log"

	"github.com/loomhq/eks-ng-ami-updater/pkg/flags"
	"github.com/loomhq/eks-ng-ami-updater/pkg/logs"
	"github.com/loomhq/eks-ng-ami-updater/pkg/updater"
)

func main() {
	debugVar, dryrunVar, skipNewerThanDays, regionsVar, nodegroupsVar, tagVar := flags.Setup()
	ctx := logs.Setup(debugVar)

	err := updater.UpdateAmi(dryrunVar, skipNewerThanDays, regionsVar, nodegroupsVar, tagVar, ctx)
	if err != nil {
		log.Fatal().Err(err).Msg("Unable to update ami")
	}
}
