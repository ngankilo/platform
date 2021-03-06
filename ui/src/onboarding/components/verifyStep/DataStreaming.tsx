// Libraries
import React, {PureComponent} from 'react'
import _ from 'lodash'

// Components
import TelegrafInstructions from 'src/onboarding/components/verifyStep/TelegrafInstructions'
import FetchConfigID from 'src/onboarding/components/verifyStep/FetchConfigID'
import FetchAuthToken from 'src/onboarding/components/verifyStep/FetchAuthToken'
import DataListening from 'src/onboarding/components/verifyStep/DataListening'

// Decorator
import {ErrorHandling} from 'src/shared/decorators/errors'

interface Props {
  bucket: string
  org: string
  username: string
}

@ErrorHandling
class DataStreaming extends PureComponent<Props> {
  public render() {
    return (
      <>
        <FetchConfigID org={this.props.org}>
          {configID => (
            <FetchAuthToken
              bucket={this.props.bucket}
              username={this.props.username}
            >
              {authToken => (
                <TelegrafInstructions
                  authToken={authToken}
                  configID={configID}
                />
              )}
            </FetchAuthToken>
          )}
        </FetchConfigID>

        <DataListening bucket={this.props.bucket} />
      </>
    )
  }
}

export default DataStreaming
