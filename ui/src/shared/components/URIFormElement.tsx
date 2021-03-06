// Libraries
import React, {PureComponent, ChangeEvent} from 'react'
import _ from 'lodash'

// Components
import {Input, ComponentStatus, FormElement} from 'src/clockface'

const VALIDATE_DEBOUNCE_MS = 350

interface Props {
  name: string
  autoFocus?: boolean
  value: string
  onChange: (e: ChangeEvent<HTMLInputElement>) => void
}

interface State {
  status: ComponentStatus
}

class URIFormElement extends PureComponent<Props, State> {
  private debouncedValidate: (value: string) => void

  constructor(props) {
    super(props)
    this.state = {
      status: ComponentStatus.Default,
    }

    this.debouncedValidate = _.debounce(this.validateURI, VALIDATE_DEBOUNCE_MS)
  }

  public render() {
    const {name, value, autoFocus} = this.props

    return (
      <FormElement label={name} key={name} errorMessage={this.errorMessage}>
        <Input
          name={name}
          autoFocus={autoFocus}
          status={this.state.status}
          onChange={this.handleChange}
          value={value}
        />
      </FormElement>
    )
  }

  private get errorMessage(): string | null {
    const {status} = this.state

    if (status === ComponentStatus.Error) {
      return 'Must be a valid URI.'
    }
  }

  private handleChange = (e: ChangeEvent<HTMLInputElement>): void => {
    const {onChange} = this.props
    const {value} = e.target

    onChange(e)
    this.debouncedValidate(value)
  }

  private validateURI = (value: string): void => {
    const regex = /http[s]?:\/\//

    if (regex.test(value)) {
      this.setState({status: ComponentStatus.Valid})
    } else {
      this.setState({status: ComponentStatus.Error})
    }
  }
}

export default URIFormElement
