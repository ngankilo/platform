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

export interface LabelType {
  id: string
  text: string
  colorHex: string
  onClick?: (id: string) => void
}

interface LabelProps {
  size?: ComponentSize
}

interface State {
  isMouseOver: boolean
}

type Props = LabelType & LabelProps

@ErrorHandling
class Label extends Component<Props, State> {
  public static defaultProps: Partial<Props> = {
    size: ComponentSize.ExtraSmall,
  }

  public static Container = LabelContainer

  constructor(props: Props) {
    super(props)

    this.state = {
      isMouseOver: false,
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
      this.setState({isMouseOver: true})
    }
  }

  private handleMouseLeave = (): void => {
    const {onClick} = this.props

    if (onClick) {
      this.setState({isMouseOver: false})
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
    const {isMouseOver} = this.state
    const {colorHex, onClick} = this.props

    let textColor
    let backgroundColor = colorHex
    const darkContrast = chroma.contrast(colorHex, Greys.Kevlar)
    const lightContrast = chroma.contrast(colorHex, Greys.White)

    if (darkContrast > lightContrast) {
      textColor = Greys.Kevlar
    } else {
      textColor = Greys.White
    }

    if (isMouseOver && onClick) {
      backgroundColor = `${chroma(colorHex).brighten(1)}`
    }

    return {
      backgroundColor: `${backgroundColor}`,
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
