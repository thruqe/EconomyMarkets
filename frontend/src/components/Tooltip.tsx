import React from 'react';

interface TooltipProps {
  content: string;
  position?: 'top' | 'bottom' | 'left' | 'right';
  children: React.ReactElement<{ className?: string }>;
  className?: string;
}

export const Tooltip: React.FC<TooltipProps> = ({
  content,
  position = 'top',
  children,
  className = '',
}) => {
  if (!content) return children;

  const posAttr = position !== 'top' ? { 'data-tooltip-pos': position } : {};

  return React.cloneElement(children, {
    'data-tooltip': content,
    ...posAttr,
    className: `${children.props.className || ''} ${className}`.trim(),
  } as any);
};
