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
import ViewTokenOverlay from './ViewTokenOverlay'

interface Props {
  auths: Authorization[]
}

interface State {
  isTokenOverlayVisible: boolean
  authInView: Authorization
}

export default class TokenList extends PureComponent<Props, State> {
  constructor(props) {
    super(props)
    this.state = {
      isTokenOverlayVisible: false,
      authInView: null,
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
              return (
                <TokenRow
                  key={a.id}
                  auth={a}
                  onClickDescription={this.handleClickDescription}
                />
              )
            })}
          </IndexList.Body>
        </IndexList>
        <OverlayTechnology visible={isTokenOverlayVisible}>
          <ViewTokenOverlay auth={this.props.auths[0]} />
        </OverlayTechnology>
      </>
    )
  }

  private handleClickDescription = (authID: string): void => {
    const authInView = this.props.auths.find(a => a.id === authID)
    this.setState({isTokenOverlayVisible: true, authInView})
  }

  private get emptyState(): JSX.Element {
    return (
      <EmptyState size={ComponentSize.Large}>
        <EmptyState.Text text="Looks no tokens match your search" />
      </EmptyState>
    )
  }
}
