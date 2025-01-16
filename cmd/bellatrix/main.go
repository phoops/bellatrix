package main

import (
	"fmt"
	"os"
	"time"

	"github.com/phoops/bellatrix/internal/core/usecases"
	"github.com/phoops/bellatrix/internal/infrastructure/state"
	"github.com/phoops/ngsiv2/client"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var (
	debugFlagName             = "debug"
	stateFileEnvVariable      = "STATE_FILE"
	debugFlagEnvVariable      = "DEBUG"
	dryRunFlagName            = "dry-run"
	dryRunEnvVariable         = "DRY_RUN"
	instancePrefixFlagName    = "instance-prefix"
	instancePrefixEnvVariable = "INSTANCE_PREFIX"
)

// Version of the program, modified by ldflags
var Version = "development"

// BuildDate is the build date of program modified by ldflags
var BuildDate = time.Now().Format("Mon Jan 2 15:04:05")

var rootCmd = &cobra.Command{
	Use:   "bellatrix",
	Short: "Bellatrix manage your orion subscriptions",
	Long:  `A fast and stateless orion subscription manger`,
}

var syncCmd = &cobra.Command{
	Run: func(cmd *cobra.Command, args []string) {
		startBellatrix(cmd, args)
	},
	Use:   "sync [CONFIG FILE]",
	Short: "Sync your orion subscriptions with your state file",
}

var versionCmd = &cobra.Command{
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Version %s, BuildDate %s", Version, BuildDate)
	},
	Use:   "version",
	Short: "Get bellatrix version",
}

func init() {
	rootCmd.PersistentFlags().Bool(debugFlagName, false, "Set the debug mode on cli")
	rootCmd.PersistentFlags().Bool(dryRunFlagName, false, "Dry run mode, does not apply patches")
	rootCmd.PersistentFlags().String(instancePrefixFlagName, "", "Optional Instance Prefix")

	rootCmd.AddCommand(syncCmd)
	rootCmd.AddCommand(versionCmd)
}

func main() {
	err := rootCmd.Execute()
	if err != nil {
		panic(err)
	}
}

func startBellatrix(cmd *cobra.Command, args []string) {
	debug, err := cmd.Flags().GetBool(debugFlagName)
	if err != nil {
		panic(err)
	}
	if !debug {
		// try for env variable
		_, debug = os.LookupEnv(debugFlagEnvVariable)
	}
	dryRun, err := cmd.Flags().GetBool(dryRunFlagName)
	if err != nil {
		panic(err)
	}
	if !dryRun {
		// try for env variable
		_, dryRun = os.LookupEnv(dryRunEnvVariable)
	}

	instancePrefix, err := cmd.Flags().GetString(instancePrefixFlagName)
	if err != nil {
		panic(err)
	}
	if len(instancePrefix) == 0 {
		// try for env variable
		instancePrefix, _ = os.LookupEnv(instancePrefixEnvVariable)
	}
	baseConfig := zap.NewDevelopmentConfig()
	var logger *zap.Logger
	if debug {
		logger, err = baseConfig.Build()
		if err != nil {
			panic(errors.Wrap(err, "could not initialize zap logger development mode"))
		}
	} else {
		baseConfig.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
		logger, err = baseConfig.Build()
		if err != nil {
			panic(errors.Wrap(err, "could not initialize zap logger production mode"))
		}
	}

	// get the state file path
	stateFilePath := ""
	if len(args) > 0 {
		stateFilePath = args[0]
	} else {
		stateFilePath = os.Getenv(stateFileEnvVariable)
	}
	if stateFilePath == "" {
		logger.Fatal("State file path not provided, aborting")
	}

	fileParser := state.NewParser(logger)
	parseSubscriptionsStateFile := usecases.NewParseSubscriptionsStateFile(
		fileParser,
		instancePrefix,
	)
	stateFromFile, err := parseSubscriptionsStateFile.Execute(stateFilePath)
	if err != nil {
		logger.Fatal("Error during state file parsing", zap.Error(err))
	}
	logger.Debug("State from file", zap.Any("content", stateFromFile))
	clientOptions := []client.ClientOptionFunc{
		client.SetUrl(stateFromFile.ClientOptions.ClientURL),
	}
	for header, value := range stateFromFile.ClientOptions.AdditionalHeaders {
		clientOptions = append(
			clientOptions,
			client.SetGlobalHeader(header, value),
		)
	}

	orionClient, err := client.NewNgsiV2Client(
		clientOptions...,
	)
	if err != nil {
		logger.Fatal("Error during orion client creation", zap.Error(err))
	}
	getAvailableSubscriptionsUsecase := usecases.NewGetAvailableSubscriptions(
		orionClient,
	)
	getSubscriptionsPatchesUsecase := usecases.NewGetSubscriptionsPatches(
		getAvailableSubscriptionsUsecase,
		logger,
		instancePrefix,
	)
	applySubscriptionsPatchesUsecase := usecases.NewApplySubscriptionsPatches(
		orionClient,
		logger,
	)
	ensureSubscriptionsAreActiveUsecase := usecases.NewEnsureSubscriptionsAreActive(
		getAvailableSubscriptionsUsecase,
		logger.Sugar(),
		orionClient,
		instancePrefix,
	)

	patches, err := getSubscriptionsPatchesUsecase.Execute(stateFromFile.SubscriptionsState)
	if err != nil {
		logger.Fatal("Error during the computing of state patches", zap.Error(err))
	}

	if !dryRun {
		err = applySubscriptionsPatchesUsecase.Execute(patches)

		if err != nil {
			logger.Fatal("Error during patch execution", zap.Error(err))
		}

		logger.Info("Ensuring the subscriptions are in the active state")
		err = ensureSubscriptionsAreActiveUsecase.Execute(
			stateFromFile.SubscriptionsState,
		)

		if err != nil {
			logger.Fatal("Error during the ensuring of subscriptions active state", zap.Error(err))
		}
	}

	logger.Info("Done, hope you had a nice sync :D")
}
