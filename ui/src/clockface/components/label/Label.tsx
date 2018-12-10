// Libraries
import React, {Component, CSSProperties} from 'react'
import chroma from 'chroma-js'
import classnames from 'classnames'

// Types
import {ComponentSize, Greys} from 'src/clockface/types'

// Components
import LabelContainer from 'src/clockface/components/label/LabelContainer'

// Styles
import './Label.scss'

import {ErrorHandling} from 'src/shared/decorators/errors'

interface Props {
  id: string
  text: string
  onClick?: (id: string) => void
  colorHex: string
  size?: ComponentSize
}

interface State {
  colorHex: string
}

@ErrorHandling
class Label extends Component<Props, State> {
  public static defaultProps: Partial<Props> = {
    size: ComponentSize.ExtraSmall,
  }

  public static Container = LabelContainer

  constructor(props: Props) {
    super(props)

    this.state = {
      colorHex: props.colorHex,
    }
  }

  public render() {
    const {text} = this.props

    this.validateColorHex()

    return (
      <label
        className={this.className}
        onClick={this.handleClick}
        onMouseEnter={this.handleMouseEnter}
        onMouseLeave={this.handleMouseLeave}
        style={this.style}
        title={this.title}
      >
        {text}
      </label>
    )
  }

  private handleClick = (): void => {
    const {id, onClick} = this.props

    if (onClick) {
      onClick(id)
    }
  }

  private handleMouseEnter = (): void => {
    const {onClick} = this.props

    if (onClick) {
      const colorHex = `${chroma(this.props.colorHex).brighten(0.85)}`
      this.setState({colorHex})
    }
  }

  private handleMouseLeave = (): void => {
    const {onClick, colorHex} = this.props

    if (onClick) {
      this.setState({colorHex})
    }
  }

  private get className(): string {
    const {size, onClick} = this.props

    return classnames('label', {
      [`label--${size}`]: size,
      'label--clickable': onClick,
    })
  }

  private get title(): string {
    const {onClick, text} = this.props

    if (onClick) {
      return `Click to see all resources with the "${text}" label`
    }
  }

  private get style(): CSSProperties {
    const {colorHex} = this.state

    let textColor = Greys.Kevlar
    const MIN_CONTRAST = 4.5
    const contrast = chroma.contrast(colorHex, textColor)

    if (contrast < MIN_CONTRAST) {
      textColor = Greys.Kevlar
    }

    return {
      backgroundColor: `${colorHex}`,
      color: `${textColor}`,
    }
  }

  private validateColorHex = (): void => {
    const {colorHex} = this.props

    const isValidLength = colorHex.length === 7
    const containsValidCharacters =
      colorHex.replace(/[ABCDEF0abcdef123456789]+/g, '') === '#'

    if (!isValidLength || !containsValidCharacters) {
      throw new Error(
        '<Label /> component has been passed a invalid hexColor prop'
      )
    }
  }
}

export default Label
