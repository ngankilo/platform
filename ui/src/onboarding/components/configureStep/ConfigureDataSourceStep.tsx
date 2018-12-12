// Libraries
import React, {PureComponent} from 'react'
import _ from 'lodash'
import {withRouter, WithRouterProps} from 'react-router'

// Components
import {ErrorHandling} from 'src/shared/decorators/errors'
import {
  Button,
  ComponentColor,
  ComponentSize,
  ComponentStatus,
} from 'src/clockface'
import ConfigureDataSourceSwitcher from 'src/onboarding/components/configureStep/ConfigureDataSourceSwitcher'

// Utils
import {
  getConfigFields,
  createNewPlugin,
} from 'src/onboarding/utils/pluginConfigs'

// Actions
import {setActiveTelegrafPlugin} from 'src/onboarding/actions/dataLoaders'
import {
  updateTelegrafPluginConfig,
  setPluginConfiguration,
  addConfigValue,
  removeConfigValue,
  createTelegrafConfigAsync,
} from 'src/onboarding/actions/dataLoaders'

// Constants
import {StepStatus} from 'src/clockface/constants/wizard'
import {
  TelegrafConfigCreationSuccess,
  TelegrafConfigCreationError,
} from 'src/shared/copy/notifications'

// Types
import {OnboardingStepProps} from 'src/onboarding/containers/OnboardingWizard'
import {
  TelegrafPlugin,
  DataLoaderType,
  ConfigurationState,
  ConfigFieldType,
  Plugin,
} from 'src/types/v2/dataLoaders'
import {validateURI} from 'src/shared/utils/validateURI'
import {getDeep} from 'src/utils/wrappers'

export interface OwnProps extends OnboardingStepProps {
  telegrafPlugins: TelegrafPlugin[]
  onSetActiveTelegrafPlugin: typeof setActiveTelegrafPlugin
  onUpdateTelegrafPluginConfig: typeof updateTelegrafPluginConfig
  onSetPluginConfiguration: typeof setPluginConfiguration
  type: DataLoaderType
  onAddConfigValue: typeof addConfigValue
  onRemoveConfigValue: typeof removeConfigValue
  onSaveTelegrafConfig: typeof createTelegrafConfigAsync
  authToken: string
}

interface RouterProps {
  params: {
    stepID: string
    substepID: string
  }
}

type Props = OwnProps & WithRouterProps & RouterProps

@ErrorHandling
class ConfigureDataSourceStep extends PureComponent<Props> {
  constructor(props: Props) {
    super(props)
  }

  public componentDidMount() {
    const {
      router,
      params: {stepID, substepID},
    } = this.props

    if (substepID === undefined) {
      router.replace(`/onboarding/${stepID}/0`)
    }
  }

  public render() {
    const {
      telegrafPlugins,
      type,
      authToken,
      params: {substepID},
      setupParams,
      onUpdateTelegrafPluginConfig,
      onSetPluginConfiguration,
      onAddConfigValue,
      onRemoveConfigValue,
    } = this.props

    return (
      <div className="onboarding-step">
        <ConfigureDataSourceSwitcher
          bucket={_.get(setupParams, 'bucket', '')}
          org={_.get(setupParams, 'org', '')}
          username={_.get(setupParams, 'username', '')}
          telegrafPlugins={telegrafPlugins}
          onUpdateTelegrafPluginConfig={onUpdateTelegrafPluginConfig}
          onSetPluginConfiguration={onSetPluginConfiguration}
          onAddConfigValue={onAddConfigValue}
          onRemoveConfigValue={onRemoveConfigValue}
          dataLoaderType={type}
          currentIndex={+substepID}
          authToken={authToken}
        />
        <div className="wizard-button-container">
          <div className="wizard-button-bar">
            <Button
              color={ComponentColor.Default}
              text="Back"
              size={ComponentSize.Medium}
              onClick={this.handlePrevious}
            />
            <Button
              color={ComponentColor.Primary}
              text="Next"
              size={ComponentSize.Medium}
              onClick={this.handleNext}
              status={ComponentStatus.Default}
              titleText={'Next'}
            />
          </div>
          {this.skipLink}
        </div>
      </div>
    )
  }

  private get skipLink() {
    return (
      <Button
        color={ComponentColor.Default}
        text="Skip"
        size={ComponentSize.Small}
        onClick={this.jumpToCompletionStep}
      >
        skip
      </Button>
    )
  }

  private jumpToCompletionStep = () => {
    const {onSetCurrentStepIndex, stepStatuses} = this.props

    onSetCurrentStepIndex(stepStatuses.length - 1)
  }

  private handleNext = async () => {
    const {
      onIncrementCurrentStepIndex,
      onSetActiveTelegrafPlugin,
      handleSetStepStatus,
      telegrafPlugins,
      authToken,
      notify,
      params: {substepID, stepID},
      router,
      type,
      onSaveTelegrafConfig,
    } = this.props

    const index = +substepID

    if (index >= telegrafPlugins.length - 1) {
      if (type === DataLoaderType.Streaming) {
        await this.setPluginConfiguration() // TODO...
        const unconfigured = this.props.telegrafPlugins.find(tp => {
          return tp.configured === ConfigurationState.Unconfigured
        })

        if (unconfigured || !telegrafPlugins.length) {
          handleSetStepStatus(parseInt(stepID, 10), StepStatus.Incomplete)
        } else {
          handleSetStepStatus(parseInt(stepID, 10), StepStatus.Complete)
        }

        try {
          await onSaveTelegrafConfig(authToken)
          notify(TelegrafConfigCreationSuccess)
        } catch (error) {
          notify(TelegrafConfigCreationError)
        }
      }

      onIncrementCurrentStepIndex()
      onSetActiveTelegrafPlugin('')
    } else {
      const name = _.get(telegrafPlugins, `${index + 1}.name`, '')
      onSetActiveTelegrafPlugin(name)

      router.push(`/onboarding/${stepID}/${index + 1}`)
    }
  }

  private handlePrevious = () => {
    const {
      router,
      onSetActiveTelegrafPlugin,
      params: {substepID},
      telegrafPlugins,
    } = this.props

    const index = +substepID

    if (index >= 0) {
      const name = _.get(telegrafPlugins, `${index - 1}.name`)
      this.setPluginConfiguration()
      onSetActiveTelegrafPlugin(name)
    } else {
      onSetActiveTelegrafPlugin('')
    }

    router.goBack()
  }

  private setPluginConfiguration = async () => {
    const {
      type,
      telegrafPlugins,
      params: {substepID},
      onSetPluginConfiguration,
    } = this.props

    const index = +substepID

    if (
      type === DataLoaderType.Streaming &&
      index <= telegrafPlugins.length - 1
    ) {
      const name = _.get(telegrafPlugins, `${index}.name`, '')
      const configFields = getConfigFields(name)

      if (!configFields) {
        onSetPluginConfiguration(name, ConfigurationState.Configured)
      } else {
        let isValidConfig = true

        const plugin = getDeep<Plugin>(
          telegrafPlugins,
          `${index}.plugin`,
          createNewPlugin(name)
        )

        const {config} = plugin

        Object.entries(configFields).forEach(configField => {
          const [fieldName, fieldType] = configField
          const fieldValue = config[fieldName]

          const isValidUri =
            fieldType === ConfigFieldType.Uri &&
            validateURI(fieldValue as string)
          const isValidString =
            fieldType === ConfigFieldType.String &&
            (fieldValue as string) !== ''
          const isValidArray =
            (fieldType === ConfigFieldType.StringArray ||
              fieldType === ConfigFieldType.UriArray) &&
            (fieldValue as string[]).length

          if (!isValidUri && !isValidString && !isValidArray) {
            isValidConfig = false
          }
        })

        if (!isValidConfig || _.isEmpty(config)) {
          onSetPluginConfiguration(name, ConfigurationState.Unconfigured)
        } else {
          onSetPluginConfiguration(name, ConfigurationState.Configured)
        }
      }
    }
  }
}

export default withRouter<OwnProps>(ConfigureDataSourceStep)
