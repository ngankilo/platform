import React, {PureComponent} from 'react'
import {Authorization} from 'src/api'

interface Props {
  auth: Authorization
}

export default class ViewTokenOverlay extends PureComponent<Props> {
  public render() {
    return <pre>{JSON.stringify(this.props.auth, null, 2)}</pre>
  }
}
