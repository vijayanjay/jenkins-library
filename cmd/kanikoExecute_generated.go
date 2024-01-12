// Code generated by piper's step-generator. DO NOT EDIT.

package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/SAP/jenkins-library/pkg/config"
	"github.com/SAP/jenkins-library/pkg/gcs"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperenv"
	"github.com/SAP/jenkins-library/pkg/splunk"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/SAP/jenkins-library/pkg/validation"
	"github.com/bmatcuk/doublestar"
	"github.com/spf13/cobra"
)

type kanikoExecuteOptions struct {
	BuildOptions                     []string                 `json:"buildOptions,omitempty"`
	BuildSettingsInfo                string                   `json:"buildSettingsInfo,omitempty"`
	ContainerBuildOptions            string                   `json:"containerBuildOptions,omitempty"`
	ContainerImage                   string                   `json:"containerImage,omitempty"`
	ContainerImageName               string                   `json:"containerImageName,omitempty" validate:"required_if=ContainerMultiImageBuild true"`
	ContainerImageTag                string                   `json:"containerImageTag,omitempty"`
	MultipleImages                   []map[string]interface{} `json:"multipleImages,omitempty"`
	ContainerMultiImageBuild         bool                     `json:"containerMultiImageBuild,omitempty"`
	ContainerMultiImageBuildExcludes []string                 `json:"containerMultiImageBuildExcludes,omitempty"`
	ContainerMultiImageBuildTrimDir  string                   `json:"containerMultiImageBuildTrimDir,omitempty"`
	ContainerPreparationCommand      string                   `json:"containerPreparationCommand,omitempty"`
	ContainerRegistryURL             string                   `json:"containerRegistryUrl,omitempty"`
	ContainerRegistryUser            string                   `json:"containerRegistryUser,omitempty"`
	ContainerRegistryPassword        string                   `json:"containerRegistryPassword,omitempty"`
	CustomTLSCertificateLinks        []string                 `json:"customTlsCertificateLinks,omitempty"`
	DockerConfigJSON                 string                   `json:"dockerConfigJSON,omitempty"`
	DockerfilePath                   string                   `json:"dockerfilePath,omitempty"`
	TargetArchitectures              []string                 `json:"targetArchitectures,omitempty"`
	ReadImageDigest                  bool                     `json:"readImageDigest,omitempty"`
	CreateBOM                        bool                     `json:"createBOM,omitempty"`
	SyftDownloadURL                  string                   `json:"syftDownloadUrl,omitempty"`
}

type kanikoExecuteCommonPipelineEnvironment struct {
	container struct {
		registryURL        string
		repositoryUsername string
		repositoryPassword string
		imageNameTag       string
		imageDigest        string
		imageNames         []string
		imageNameTags      []string
		imageDigests       []string
	}
	custom struct {
		buildSettingsInfo string
	}
}

func (p *kanikoExecuteCommonPipelineEnvironment) persist(path, resourceName string) {
	content := []struct {
		category string
		name     string
		value    interface{}
	}{
		{category: "container", name: "registryUrl", value: p.container.registryURL},
		{category: "container", name: "repositoryUsername", value: p.container.repositoryUsername},
		{category: "container", name: "repositoryPassword", value: p.container.repositoryPassword},
		{category: "container", name: "imageNameTag", value: p.container.imageNameTag},
		{category: "container", name: "imageDigest", value: p.container.imageDigest},
		{category: "container", name: "imageNames", value: p.container.imageNames},
		{category: "container", name: "imageNameTags", value: p.container.imageNameTags},
		{category: "container", name: "imageDigests", value: p.container.imageDigests},
		{category: "custom", name: "buildSettingsInfo", value: p.custom.buildSettingsInfo},
	}

	errCount := 0
	for _, param := range content {
		err := piperenv.SetResourceParameter(path, resourceName, filepath.Join(param.category, param.name), param.value)
		if err != nil {
			log.Entry().WithError(err).Error("Error persisting piper environment.")
			errCount++
		}
	}
	if errCount > 0 {
		log.Entry().Error("failed to persist Piper environment")
	}
}

type kanikoExecuteReports struct {
}

func (p *kanikoExecuteReports) persist(stepConfig kanikoExecuteOptions, gcpJsonKeyFilePath string, gcsBucketId string, gcsFolderPath string, gcsSubFolder string) {
	if gcsBucketId == "" {
		log.Entry().Info("persisting reports to GCS is disabled, because gcsBucketId is empty")
		return
	}
	log.Entry().Info("Uploading reports to Google Cloud Storage...")
	content := []gcs.ReportOutputParam{
		{FilePattern: "**/bom-*.xml", ParamRef: "", StepResultType: "sbom"},
	}
	envVars := []gcs.EnvVar{
		{Name: "GOOGLE_APPLICATION_CREDENTIALS", Value: gcpJsonKeyFilePath, Modified: false},
	}
	gcsClient, err := gcs.NewClient(gcs.WithEnvVars(envVars))
	if err != nil {
		log.Entry().Errorf("creation of GCS client failed: %v", err)
		return
	}
	defer gcsClient.Close()
	structVal := reflect.ValueOf(&stepConfig).Elem()
	inputParameters := map[string]string{}
	for i := 0; i < structVal.NumField(); i++ {
		field := structVal.Type().Field(i)
		if field.Type.String() == "string" {
			paramName := strings.Split(field.Tag.Get("json"), ",")
			paramValue, _ := structVal.Field(i).Interface().(string)
			inputParameters[paramName[0]] = paramValue
		}
	}
	if err := gcs.PersistReportsToGCS(gcsClient, content, inputParameters, gcsFolderPath, gcsBucketId, gcsSubFolder, doublestar.Glob, os.Stat); err != nil {
		log.Entry().Errorf("failed to persist reports: %v", err)
	}
}

// KanikoExecuteCommand Executes a [Kaniko](https://github.com/GoogleContainerTools/kaniko) build for creating a Docker container.
func KanikoExecuteCommand() *cobra.Command {
	const STEP_NAME = "kanikoExecute"

	metadata := kanikoExecuteMetadata()
	var stepConfig kanikoExecuteOptions
	var startTime time.Time
	var commonPipelineEnvironment kanikoExecuteCommonPipelineEnvironment
	var reports kanikoExecuteReports
	var logCollector *log.CollectorHook
	var splunkClient *splunk.Splunk
	telemetryClient := &telemetry.Telemetry{}

	var createKanikoExecuteCmd = &cobra.Command{
		Use:   STEP_NAME,
		Short: "Executes a [Kaniko](https://github.com/GoogleContainerTools/kaniko) build for creating a Docker container.",
		Long: `Executes a [Kaniko](https://github.com/GoogleContainerTools/kaniko) build for creating a Docker container.

### Building one container image

For building one container image the step expects that one of the containerImage, containerImageName or --destination (via buildOptions) is set.

### Building multiple container images

The step allows you to build multiple container images with one run.
This is suitable in case you need to create multiple images for one microservice, e.g. for testing.

All images will get the same "root" name and the same versioning.<br />
**Thus, this is not suitable to be used for a monorepo approach!** For monorepos you need to use a build tool natively capable to take care for monorepos
or implement a custom logic and for example execute this ` + "`" + `kanikoExecute` + "`" + ` step multiple times in your custom pipeline.

You can activate multiple builds using the parameter [containerMultiImageBuild](#containermultiimagebuild)

Behavior can be adapted using:

* [containerMultiImageBuildExcludes](#containermultiimagebuildexcludes) for defining excludes
* [containerMultiImageBuildTrimDir](#containermultiimagebuildtrimdir) for removing parent directory part from image name

Examples:

#### Multiple containers in sub directories

Configuration as follows:

` + "`" + `` + "`" + `` + "`" + `
general:
  containerImageName: myImage
steps:
  kanikoExecute:
    containerMultiImageBuild: true
` + "`" + `` + "`" + `` + "`" + `

Following Dockerfiles are available in the repository:

* sub1/Dockerfile
* sub2/Dockerfile

Following final image names will be built:

* ` + "`" + `myImage-sub1` + "`" + `
* ` + "`" + `myImage-sub2` + "`" + `

#### Multiple containers in sub directories while trimming a directory part

Configuration as follows:

` + "`" + `` + "`" + `` + "`" + `
general:
  containerImageName: myImage
steps:
  kanikoExecute:
    containerMultiImageBuild: true
    containerMultiImageBuildTrimDir: .ci
` + "`" + `` + "`" + `` + "`" + `

Following Dockerfiles are available in the repository:

* .ci/sub1/Dockerfile
* .ci/sub2/Dockerfile

Following final image names will be built:

* ` + "`" + `myImage-sub1` + "`" + `
* ` + "`" + `myImage-sub2` + "`" + ``,
		PreRunE: func(cmd *cobra.Command, _ []string) error {
			startTime = time.Now()
			log.SetStepName(STEP_NAME)
			log.SetVerbose(GeneralConfig.Verbose)

			GeneralConfig.GitHubAccessTokens = ResolveAccessTokens(GeneralConfig.GitHubTokens)

			path, _ := os.Getwd()
			fatalHook := &log.FatalHook{CorrelationID: GeneralConfig.CorrelationID, Path: path}
			log.RegisterHook(fatalHook)

			err := PrepareConfig(cmd, &metadata, STEP_NAME, &stepConfig, config.OpenPiperFile)
			if err != nil {
				log.SetErrorCategory(log.ErrorConfiguration)
				return err
			}
			log.RegisterSecret(stepConfig.DockerConfigJSON)

			if len(GeneralConfig.HookConfig.SentryConfig.Dsn) > 0 {
				sentryHook := log.NewSentryHook(GeneralConfig.HookConfig.SentryConfig.Dsn, GeneralConfig.CorrelationID)
				log.RegisterHook(&sentryHook)
			}

			if len(GeneralConfig.HookConfig.SplunkConfig.Dsn) > 0 || len(GeneralConfig.HookConfig.SplunkConfig.ProdCriblEndpoint) > 0 {
				splunkClient = &splunk.Splunk{}
				logCollector = &log.CollectorHook{CorrelationID: GeneralConfig.CorrelationID}
				log.RegisterHook(logCollector)
			}

			if err = log.RegisterANSHookIfConfigured(GeneralConfig.CorrelationID); err != nil {
				log.Entry().WithError(err).Warn("failed to set up SAP Alert Notification Service log hook")
			}

			validation, err := validation.New(validation.WithJSONNamesForStructFields(), validation.WithPredefinedErrorMessages())
			if err != nil {
				return err
			}
			if err = validation.ValidateStruct(stepConfig); err != nil {
				log.SetErrorCategory(log.ErrorConfiguration)
				return err
			}

			return nil
		},
		Run: func(_ *cobra.Command, _ []string) {
			stepTelemetryData := telemetry.CustomData{}
			stepTelemetryData.ErrorCode = "1"
			handler := func() {
				commonPipelineEnvironment.persist(GeneralConfig.EnvRootPath, "commonPipelineEnvironment")
				reports.persist(stepConfig, GeneralConfig.GCPJsonKeyFilePath, GeneralConfig.GCSBucketId, GeneralConfig.GCSFolderPath, GeneralConfig.GCSSubFolder)
				config.RemoveVaultSecretFiles()
				stepTelemetryData.Duration = fmt.Sprintf("%v", time.Since(startTime).Milliseconds())
				stepTelemetryData.ErrorCategory = log.GetErrorCategory().String()
				stepTelemetryData.PiperCommitHash = GitCommit
				telemetryClient.SetData(&stepTelemetryData)
				telemetryClient.Send()
				if len(GeneralConfig.HookConfig.SplunkConfig.Dsn) > 0 {
					splunkClient.Initialize(GeneralConfig.CorrelationID,
						GeneralConfig.HookConfig.SplunkConfig.Dsn,
						GeneralConfig.HookConfig.SplunkConfig.Token,
						GeneralConfig.HookConfig.SplunkConfig.Index,
						GeneralConfig.HookConfig.SplunkConfig.SendLogs)
					splunkClient.Send(telemetryClient.GetData(), logCollector)
				}
				if len(GeneralConfig.HookConfig.SplunkConfig.ProdCriblEndpoint) > 0 {
					splunkClient.Initialize(GeneralConfig.CorrelationID,
						GeneralConfig.HookConfig.SplunkConfig.ProdCriblEndpoint,
						GeneralConfig.HookConfig.SplunkConfig.ProdCriblToken,
						GeneralConfig.HookConfig.SplunkConfig.ProdCriblIndex,
						GeneralConfig.HookConfig.SplunkConfig.SendLogs)
					splunkClient.Send(telemetryClient.GetData(), logCollector)
				}
			}
			log.DeferExitHandler(handler)
			defer handler()
			telemetryClient.Initialize(GeneralConfig.NoTelemetry, STEP_NAME)
			kanikoExecute(stepConfig, &stepTelemetryData, &commonPipelineEnvironment)
			stepTelemetryData.ErrorCode = "0"
			log.Entry().Info("SUCCESS")
		},
	}

	addKanikoExecuteFlags(createKanikoExecuteCmd, &stepConfig)
	return createKanikoExecuteCmd
}

func addKanikoExecuteFlags(cmd *cobra.Command, stepConfig *kanikoExecuteOptions) {
	cmd.Flags().StringSliceVar(&stepConfig.BuildOptions, "buildOptions", []string{`--skip-tls-verify-pull`, `--ignore-path=/workspace`, `--ignore-path=/busybox`}, "Defines a list of build options for the [kaniko](https://github.com/GoogleContainerTools/kaniko) build.")
	cmd.Flags().StringVar(&stepConfig.BuildSettingsInfo, "buildSettingsInfo", os.Getenv("PIPER_buildSettingsInfo"), "Build settings info is typically filled by the step automatically to create information about the build settings that were used during the mta build. This information is typically used for compliance related processes.")
	cmd.Flags().StringVar(&stepConfig.ContainerBuildOptions, "containerBuildOptions", os.Getenv("PIPER_containerBuildOptions"), "Deprected, please use buildOptions. Defines the build options for the [kaniko](https://github.com/GoogleContainerTools/kaniko) build.")
	cmd.Flags().StringVar(&stepConfig.ContainerImage, "containerImage", os.Getenv("PIPER_containerImage"), "Defines the full name of the Docker image to be created including registry, image name and tag like `my.docker.registry/path/myImageName:myTag`. If `containerImage` is not provided, then `containerImageName` or `--destination` (via buildOptions) should be provided.")
	cmd.Flags().StringVar(&stepConfig.ContainerImageName, "containerImageName", os.Getenv("PIPER_containerImageName"), "Name of the container which will be built - will be used instead of parameter `containerImage`. If `containerImageName` is not provided, then `containerImage` or `--destination` (via buildOptions) should be provided.")
	cmd.Flags().StringVar(&stepConfig.ContainerImageTag, "containerImageTag", os.Getenv("PIPER_containerImageTag"), "Tag of the container which will be built - will be used instead of parameter `containerImage`")

	cmd.Flags().BoolVar(&stepConfig.ContainerMultiImageBuild, "containerMultiImageBuild", false, "Defines if multiple containers should be build. Dockerfiles are used using the pattern **/Dockerfile*. Excludes can be defined via [`containerMultiImageBuildExcludes`](#containermultiimagebuildexscludes).")
	cmd.Flags().StringSliceVar(&stepConfig.ContainerMultiImageBuildExcludes, "containerMultiImageBuildExcludes", []string{}, "Defines a list of Dockerfile paths to exclude from the build when using [`containerMultiImageBuild`](#containermultiimagebuild).")
	cmd.Flags().StringVar(&stepConfig.ContainerMultiImageBuildTrimDir, "containerMultiImageBuildTrimDir", os.Getenv("PIPER_containerMultiImageBuildTrimDir"), "Defines a trailing directory part which should not be considered in the final image name.")
	cmd.Flags().StringVar(&stepConfig.ContainerPreparationCommand, "containerPreparationCommand", `rm -f /kaniko/.docker/config.json`, "Defines the command to prepare the Kaniko container. By default the contained credentials are removed in order to allow anonymous access to container registries.")
	cmd.Flags().StringVar(&stepConfig.ContainerRegistryURL, "containerRegistryUrl", os.Getenv("PIPER_containerRegistryUrl"), "http(s) url of the Container registry where the image should be pushed to - will be used instead of parameter `containerImage`")
	cmd.Flags().StringVar(&stepConfig.ContainerRegistryUser, "containerRegistryUser", os.Getenv("PIPER_containerRegistryUser"), "Username of the Container registry where the image should be pushed to - which will updated in a docker config json file. If a docker config json file is provided via parameter `dockerConfigJSON` , then the existing file will be enhanced")
	cmd.Flags().StringVar(&stepConfig.ContainerRegistryPassword, "containerRegistryPassword", os.Getenv("PIPER_containerRegistryPassword"), "Password of the Container registry where the image should be pushed to -  which will updated in a docker config json file. If a docker config json file is provided via parameter `dockerConfigJSON` , then the existing file will be enhanced")
	cmd.Flags().StringSliceVar(&stepConfig.CustomTLSCertificateLinks, "customTlsCertificateLinks", []string{}, "List containing download links of custom TLS certificates. This is required to ensure trusted connections to registries with custom certificates.")
	cmd.Flags().StringVar(&stepConfig.DockerConfigJSON, "dockerConfigJSON", os.Getenv("PIPER_dockerConfigJSON"), "Path to the file `.docker/config.json` - this is typically provided by your CI/CD system. You can find more details about the Docker credentials in the [Docker documentation](https://docs.docker.com/engine/reference/commandline/login/).")
	cmd.Flags().StringVar(&stepConfig.DockerfilePath, "dockerfilePath", `Dockerfile`, "Defines the location of the Dockerfile relative to the Jenkins workspace.")
	cmd.Flags().StringSliceVar(&stepConfig.TargetArchitectures, "targetArchitectures", []string{``}, "Defines the target architectures for which the build should run using OS and architecture separated by a comma. (EXPERIMENTAL)")
	cmd.Flags().BoolVar(&stepConfig.ReadImageDigest, "readImageDigest", false, "")
	cmd.Flags().BoolVar(&stepConfig.CreateBOM, "createBOM", false, "Creates the bill of materials (BOM) using Syft and stores it in a file in CycloneDX 1.4 format.")
	cmd.Flags().StringVar(&stepConfig.SyftDownloadURL, "syftDownloadUrl", `https://github.com/anchore/syft/releases/download/v0.62.3/syft_0.62.3_linux_amd64.tar.gz`, "Specifies the download url of the Syft Linux amd64 tar binary file. This can be found at https://github.com/anchore/syft/releases/.")

}

// retrieve step metadata
func kanikoExecuteMetadata() config.StepData {
	var theMetaData = config.StepData{
		Metadata: config.StepMetadata{
			Name:        "kanikoExecute",
			Aliases:     []config.Alias{},
			Description: "Executes a [Kaniko](https://github.com/GoogleContainerTools/kaniko) build for creating a Docker container.",
		},
		Spec: config.StepSpec{
			Inputs: config.StepInputs{
				Secrets: []config.StepSecrets{
					{Name: "dockerConfigJsonCredentialsId", Description: "Jenkins 'Secret file' credentials ID containing Docker config.json (with registry credential(s)). You can create it like explained in the [protocodeExecuteScan Prerequisites section](https://www.project-piper.io/steps/protecodeExecuteScan/#prerequisites).", Type: "jenkins"},
				},
				Parameters: []config.StepParameters{
					{
						Name:        "buildOptions",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "[]string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     []string{`--skip-tls-verify-pull`, `--ignore-path=/workspace`, `--ignore-path=/busybox`},
					},
					{
						Name: "buildSettingsInfo",
						ResourceRef: []config.ResourceReference{
							{
								Name:  "commonPipelineEnvironment",
								Param: "custom/buildSettingsInfo",
							},
						},
						Scope:     []string{"STEPS", "STAGES", "PARAMETERS"},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{},
						Default:   os.Getenv("PIPER_buildSettingsInfo"),
					},
					{
						Name:        "containerBuildOptions",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_containerBuildOptions"),
					},
					{
						Name:        "containerImage",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "containerImageNameAndTag", Deprecated: true}},
						Default:     os.Getenv("PIPER_containerImage"),
					},
					{
						Name:        "containerImageName",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "dockerImageName"}},
						Default:     os.Getenv("PIPER_containerImageName"),
					},
					{
						Name: "containerImageTag",
						ResourceRef: []config.ResourceReference{
							{
								Name:  "commonPipelineEnvironment",
								Param: "artifactVersion",
							},
						},
						Scope:     []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{{Name: "artifactVersion"}},
						Default:   os.Getenv("PIPER_containerImageTag"),
					},
					{
						Name:        "multipleImages",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STEPS"},
						Type:        "[]map[string]interface{}",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "images"}},
					},
					{
						Name:        "containerMultiImageBuild",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:        "bool",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     false,
					},
					{
						Name:        "containerMultiImageBuildExcludes",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:        "[]string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     []string{},
					},
					{
						Name:        "containerMultiImageBuildTrimDir",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_containerMultiImageBuildTrimDir"),
					},
					{
						Name:        "containerPreparationCommand",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     `rm -f /kaniko/.docker/config.json`,
					},
					{
						Name: "containerRegistryUrl",
						ResourceRef: []config.ResourceReference{
							{
								Name:  "commonPipelineEnvironment",
								Param: "container/registryUrl",
							},
						},
						Scope:     []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{{Name: "dockerRegistryUrl"}},
						Default:   os.Getenv("PIPER_containerRegistryUrl"),
					},
					{
						Name: "containerRegistryUser",
						ResourceRef: []config.ResourceReference{
							{
								Name:  "commonPipelineEnvironment",
								Param: "container/repositoryUsername",
							},
						},
						Scope:     []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{{Name: "dockerRegistryUser"}},
						Default:   os.Getenv("PIPER_containerRegistryUser"),
					},
					{
						Name: "containerRegistryPassword",
						ResourceRef: []config.ResourceReference{
							{
								Name:  "commonPipelineEnvironment",
								Param: "container/repositoryPassword",
							},
						},
						Scope:     []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{{Name: "dockerRegistryPassword"}},
						Default:   os.Getenv("PIPER_containerRegistryPassword"),
					},
					{
						Name:        "customTlsCertificateLinks",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "[]string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     []string{},
					},
					{
						Name: "dockerConfigJSON",
						ResourceRef: []config.ResourceReference{
							{
								Name: "dockerConfigJsonCredentialsId",
								Type: "secret",
							},

							{
								Name:    "dockerConfigFileVaultSecretName",
								Type:    "vaultSecretFile",
								Default: "docker-config",
							},
						},
						Scope:     []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: false,
						Aliases:   []config.Alias{},
						Default:   os.Getenv("PIPER_dockerConfigJSON"),
					},
					{
						Name:        "dockerfilePath",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{{Name: "dockerfile"}},
						Default:     `Dockerfile`,
					},
					{
						Name:        "targetArchitectures",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"GENERAL", "STEPS", "STAGES", "PARAMETERS"},
						Type:        "[]string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     []string{``},
					},
					{
						Name:        "readImageDigest",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"STEPS", "STAGES", "PARAMETERS"},
						Type:        "bool",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     false,
					},
					{
						Name:        "createBOM",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"GENERAL", "STEPS", "STAGES", "PARAMETERS"},
						Type:        "bool",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     false,
					},
					{
						Name:        "syftDownloadUrl",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     `https://github.com/anchore/syft/releases/download/v0.62.3/syft_0.62.3_linux_amd64.tar.gz`,
					},
				},
			},
			Containers: []config.Container{
				{Image: "gcr.io/kaniko-project/executor:debug", EnvVars: []config.EnvVar{{Name: "container", Value: "docker"}}, Options: []config.Option{{Name: "-u", Value: "0"}, {Name: "--entrypoint", Value: ""}}},
			},
			Outputs: config.StepOutputs{
				Resources: []config.StepResources{
					{
						Name: "commonPipelineEnvironment",
						Type: "piperEnvironment",
						Parameters: []map[string]interface{}{
							{"name": "container/registryUrl"},
							{"name": "container/repositoryUsername"},
							{"name": "container/repositoryPassword"},
							{"name": "container/imageNameTag"},
							{"name": "container/imageDigest"},
							{"name": "container/imageNames", "type": "[]string"},
							{"name": "container/imageNameTags", "type": "[]string"},
							{"name": "container/imageDigests", "type": "[]string"},
							{"name": "custom/buildSettingsInfo"},
						},
					},
					{
						Name: "reports",
						Type: "reports",
						Parameters: []map[string]interface{}{
							{"filePattern": "**/bom-*.xml", "type": "sbom"},
						},
					},
				},
			},
		},
	}
	return theMetaData
}
