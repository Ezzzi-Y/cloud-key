import { useState, useEffect } from 'react'
import { format } from 'date-fns'
import { zhCN } from 'date-fns/locale/zh-CN'
import { CalendarIcon, X } from 'lucide-react'
import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import { Calendar } from '@/components/ui/calendar'
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover'

interface DateTimePickerProps {
  value: string // "YYYY-MM-DDTHH:mm" format
  onChange: (val: string) => void
  placeholder?: string
  className?: string
  clearable?: boolean
}

function parseValue(val: string): { date: Date | undefined; time: string } {
  if (!val) return { date: undefined, time: '' }
  const [datePart, timePart] = val.split('T')
  if (!datePart) return { date: undefined, time: '' }
  const date = new Date(datePart + 'T00:00:00')
  return { date: isNaN(date.getTime()) ? undefined : date, time: timePart || '' }
}

function formatValue(date: Date | undefined, time: string): string {
  if (!date) return ''
  const yyyy = date.getFullYear()
  const mm = String(date.getMonth() + 1).padStart(2, '0')
  const dd = String(date.getDate()).padStart(2, '0')
  return `${yyyy}-${mm}-${dd}T${time || '00:00'}`
}

export function DateTimePicker({
  value,
  onChange,
  placeholder = '选择时间',
  className,
  clearable = true,
}: DateTimePickerProps) {
  const [open, setOpen] = useState(false)
  const { date, time } = parseValue(value)
  const [tempDate, setTempDate] = useState<Date | undefined>(date)
  const [tempTime, setTempTime] = useState(time)

  // Sync when value changes externally
  useEffect(() => {
    const parsed = parseValue(value)
    setTempDate(parsed.date)
    setTempTime(parsed.time)
  }, [value])

  const handleSelect = (selected: Date | undefined) => {
    setTempDate(selected)
    if (selected) {
      onChange(formatValue(selected, tempTime || '00:00'))
    }
  }

  const handleTimeChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const newTime = e.target.value
    setTempTime(newTime)
    if (tempDate) {
      onChange(formatValue(tempDate, newTime))
    }
  }

  const handleClear = (e: React.MouseEvent) => {
    e.stopPropagation()
    onChange('')
    setTempDate(undefined)
    setTempTime('')
  }

  const displayText = value
    ? `${format(date!, 'yyyy-MM-dd')} ${time || '00:00'}`
    : null

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <Button
          variant="outline"
          className={cn(
            'w-full max-w-[200px] justify-start text-left font-normal',
            !value && 'text-muted-foreground',
            className
          )}
        >
          <CalendarIcon className="mr-2 h-4 w-4 shrink-0" />
          <span className="flex-1 truncate">
            {displayText || placeholder}
          </span>
          {clearable && value && (
            <X
              className="ml-1 h-3.5 w-3.5 shrink-0 opacity-50 hover:opacity-100"
              onClick={handleClear}
            />
          )}
        </Button>
      </PopoverTrigger>
      <PopoverContent className="w-auto p-0" align="start">
        <Calendar
          mode="single"
          selected={tempDate}
          onSelect={handleSelect}
          locale={zhCN}
        />
        <div className="border-t px-3 py-2">
          <label className="flex items-center gap-2 text-sm text-muted-foreground">
            时间
            <input
              type="time"
              value={tempTime}
              onChange={handleTimeChange}
              className="rounded border border-input bg-background px-2 py-1 text-sm text-foreground focus:outline-none focus:ring-2 focus:ring-ring"
            />
          </label>
        </div>
      </PopoverContent>
    </Popover>
  )
}
