// Libraries
import React, {Component, ChangeEvent, KeyboardEvent} from 'react'
import classnames from 'classnames'
import _ from 'lodash'

// Types
// import {ComponentSize} from 'src/clockface/types'

// Components
import Input from 'src/clockface/components/inputs/Input'
import Label, {LabelType} from 'src/clockface/components/label/Label'
import LabelContainer from 'src/clockface/components/label/LabelContainer'

// Styles
import './LabelSelector.scss'

import {ErrorHandling} from 'src/shared/decorators/errors'

interface Props {
  selectedLabels: LabelType[]
  checkAgainst: LabelType[]
  onAddLabel: (label: LabelType) => void
  onRemoveLabel: (label: LabelType) => void
  resourceType: string
}

interface State {
  filterValue: string
  isSuggesting: boolean
  highlightedID: string
}

@ErrorHandling
class LabelSelector extends Component<Props, State> {
  constructor(props: Props) {
    super(props)

    this.state = {
      highlightedID: null,
      filterValue: '',
      isSuggesting: false,
    }
  }

  public render() {
    const {resourceType} = this.props
    const {filterValue} = this.state

    return (
      <div className="label-selector">
        <div className="label-selector--input">
          <Input
            placeholder={`Add labels to ${resourceType}`}
            value={filterValue}
            onFocus={this.handleStartSuggesting}
            onKeyDown={this.handleKeyDown}
            onChange={this.handleInputChange}
          />
          {this.suggestionMenu}
        </div>
        {this.selectedLabels}
      </div>
    )
  }

  private get selectedLabels(): JSX.Element {
    const {selectedLabels, resourceType} = this.props

    if (selectedLabels && selectedLabels.length) {
      return (
        <LabelContainer className="label-selector--selected">
          {selectedLabels.map(label => (
            <Label
              key={label.id}
              text={label.text}
              id={label.id}
              colorHex={label.colorHex}
              onDelete={this.handleDelete}
            />
          ))}
        </LabelContainer>
      )
    }

    return <p>{`This ${resourceType} has no labels`}</p>
  }

  private get availableLabels(): LabelType[] {
    const {selectedLabels, checkAgainst} = this.props

    return _.difference(checkAgainst, [...selectedLabels])
  }

  private get filteredLabels(): LabelType[] {
    const {filterValue} = this.state

    return _.filter(this.props.checkAgainst, l => {
      const itemText = _.lowerCase(l.text)

      return itemText.includes(_.lowerCase(filterValue))
    })
  }

  private get suggestionMenu(): JSX.Element {
    const {isSuggesting} = this.state

    if (isSuggesting) {
      return (
        <div>
          {this.filteredLabels.length ? (
            this.filteredLabels.map(item => {
              return (
                <div
                  key={item.id}
                  className={this.suggestionItemClass(item.id)}
                  onMouseOver={this.handleSetHighlightedID(item.id)}
                  onClick={this.handleSuggestionClick(item.id)}
                >
                  {item.text}
                </div>
              )
            })
          ) : (
            <div>No matching labels</div>
          )}
        </div>
      )
    }
  }

  private suggestionItemClass = (labelID: string): string => {
    const {highlightedID} = this.state

    return classnames('suggest-item', {highlight: labelID === highlightedID})
  }

  private handleSetHighlightedID = (highlightedID: string) => (): void => {
    this.setState({highlightedID})
  }

  private handleStartSuggesting = () => {
    const highlightedID = this.props.checkAgainst[0].id
    this.setState({isSuggesting: true, highlightedID})
  }

  private handleStopSuggesting = () => {
    this.setState({isSuggesting: false, highlightedID: null})
  }

  private handleInputChange = (e: ChangeEvent<HTMLInputElement>): void => {
    const {highlightedID} = this.state
    const filterValue = e.target.value

    const filteredItems = _.filter(this.availableLabels, label => {
      const itemText = _.lowerCase(label.text)

      return itemText.includes(_.lowerCase(filterValue))
    })

    if (filteredItems.find(item => item.id === highlightedID)) {
      this.setState({filterValue})
    } else if (filteredItems.length) {
      const highlightedID = filteredItems[0].id
      this.setState({filterValue, highlightedID})
    } else {
      this.setState({filterValue})
    }
  }

  private handleDelete = (labelID: string): void => {
    const {onRemoveLabel, selectedLabels} = this.props

    const label = selectedLabels.find(l => l.id === labelID)

    onRemoveLabel(label)
  }

  private handleSuggestionClick = (labelID: string) => (): void => {
    const {onAddLabel} = this.props

    const label = this.props.checkAgainst.find(label => label.id === labelID)

    onAddLabel(label)
    this.handleStopSuggesting()
  }

  private handleKeyDown = (e: KeyboardEvent<HTMLInputElement>): void => {
    if (e.key === 'Escape') {
      console.log('escape!')
      this.handleStopSuggesting()
    } else if (e.key === 'Enter') {
      const label = this.props.checkAgainst.find(
        label => label.id === this.state.highlightedID
      )

      this.props.onAddLabel(label)
    }
  }
}

export default LabelSelector
