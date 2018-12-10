// Libraries
import React, {PureComponent} from 'react'

// Components
import {
  IndexList,
  EmptyState,
  ComponentSize,
  OverlayTechnology,
} from 'src/clockface'
import TokenRow from 'src/me/components/account/TokenRow'

// Types
import {Authorization} from 'src/api'

interface Props {
  auths: Authorization[]
}

interface State {
  isTokenOverlayVisible: boolean
}

export default class TokenList extends PureComponent<Props, State> {
  constructor(props) {
    super(props)
    this.state = {
      isTokenOverlayVisible: false,
    }
  }

  public render() {
    const {auths} = this.props
    const {isTokenOverlayVisible} = this.state

    return (
      <>
        <IndexList>
          <IndexList.Header>
            <IndexList.HeaderCell columnName="Description" />
            <IndexList.HeaderCell columnName="Status" />
          </IndexList.Header>
          <IndexList.Body emptyState={this.emptyState} columnCount={2}>
            {auths.map(a => {
              return <TokenRow key={a.id} auth={a} />
            })}
          </IndexList.Body>
        </IndexList>
        <OverlayTechnology visible={isTokenOverlayVisible}>
          <div>hai im an overlay</div>
        </OverlayTechnology>
      </>
    )
  }

  private get emptyState(): JSX.Element {
    return (
      <EmptyState size={ComponentSize.Large}>
        <EmptyState.Text text="Looks no tokens match your search" />
      </EmptyState>
    )
  }
}
