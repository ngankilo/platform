// Libraries
import React, {SFC} from 'react'
import classnames from 'classnames'

interface Props {
  children: JSX.Element | JSX.Element[]
  className?: string
}

const LabelContainer: SFC<Props> = ({children, className}) => (
  <div
    className={classnames('label--container', {[`${className}`]: className})}
  >
    <div className="label--container-margin">{children}</div>
  </div>
)

export default LabelContainer
