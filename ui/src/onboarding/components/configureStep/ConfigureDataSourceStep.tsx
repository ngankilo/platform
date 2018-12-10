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
import {getConfigFields} from 'src/onboarding/utils/pluginConfigs'

// Actions
import {setActiveTelegrafPlugin} from 'src/onboarding/actions/dataLoaders'
import {
  updateTelegrafPluginConfig,
  setPluginConfigurationState,
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
} from 'src/types/v2/dataLoaders'
import {validateURI} from 'src/shared/utils/validateURI'

export interface OwnProps extends OnboardingStepProps {
  telegrafPlugins: TelegrafPlugin[]
  onSetActiveTelegrafPlugin: typeof setActiveTelegrafPlugin
  onUpdateTelegrafPluginConfig: typeof updateTelegrafPluginConfig
  onSetPluginConfigurationState: typeof setPluginConfigurationState
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
      onSetPluginConfigurationState,
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
          onSetPluginConfigurationState={onSetPluginConfigurationState}
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

    await this.setPluginConfiguration()

    if (index >= telegrafPlugins.length - 1) {
      if (type === DataLoaderType.Streaming) {
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
      onSetPluginConfigurationState,
    } = this.props

    const index = +substepID

    if (
      type === DataLoaderType.Streaming &&
      index <= telegrafPlugins.length - 1
    ) {
      const name = _.get(telegrafPlugins, `${index}.name`, '')
      const configFields = getConfigFields(name)

      if (!configFields) {
        onSetPluginConfigurationState(name, ConfigurationState.Configured)
      } else {
        let invalidConfigField = false
        const config = _.get(telegrafPlugins, `${index}.plugin.config`, {})

        Object.keys(config).forEach(fieldName => {
          const fieldValue = config[fieldName]

          if (fieldName === 'url' && !validateURI(fieldValue)) {
            invalidConfigField = true
          } else if (fieldValue === '') {
            invalidConfigField = true
          }
        })

        if (invalidConfigField) {
          onSetPluginConfigurationState(name, ConfigurationState.Unconfigured)
        } else {
          onSetPluginConfigurationState(name, ConfigurationState.Configured)
        }
      }
    }
  }
}

export default withRouter<OwnProps>(ConfigureDataSourceStep)
