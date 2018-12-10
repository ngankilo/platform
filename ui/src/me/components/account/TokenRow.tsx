// Libraries
import React, {PureComponent} from 'react'

// Components
import {IndexList} from 'src/clockface'

// Types
import {Authorization} from 'src/api'

interface Props {
  auth: Authorization
  onClickDescription?: (authID: string) => void
}

export default class TokenRow extends PureComponent<Props> {
  public render() {
    const {description, status} = this.props.auth

    return (
      <IndexList.Row>
        <IndexList.Cell>
          <span onClick={this.handleClickDescription}>{description}</span>
        </IndexList.Cell>
        <IndexList.Cell>{status}</IndexList.Cell>
      </IndexList.Row>
    )
  }

  private handleClickDescription = () => {
    // const {onClickDescription, auth} = this.props
    // onClickDescription(auth.id)
  }
}
