import {
  ActionIcon,
  Blockquote,
  Box,
  Card,
  Code,
  CopyButton,
  Divider,
  Group,
  List,
  Mark,
  Paper,
  ScrollArea,
  Stack,
  Table,
  Text,
  ThemeIcon,
  Title,
  Tooltip,
} from '@mantine/core'
import {
  IconBrain,
  IconCheck,
  IconCopy,
  IconFileText,
  IconInfoCircle,
  IconQuote,
  IconTerminal2,
} from '@tabler/icons-react'
import React, { useMemo } from 'react'

// Type definitions for parsed content
type ParsedElement =
  | {
      type: 'h1' | 'h2' | 'h3' | 'text' | 'blockquote' | 'table-row'
      content: string
      key: string | number
    }
  | { type: 'code'; content: string; key: string | number }
  | {
      type: 'codeblock'
      content: string
      language: string
      key: string | number
    }
  | { type: 'list-item' | 'ordered-list-item'; content: string; key: string }

type GroupedElement =
  | ParsedElement
  | { type: 'unordered' | 'ordered'; items: string[]; key?: string }

// Simple markdown parser for common patterns
function parseMarkdown(text: string): ParsedElement[] {
  // Split by code blocks first
  const parts = text.split(/(```[\s\S]*?```|`[^`]+`)/g)

  return parts
    .map((part, index) => {
      // Code blocks
      if (part.startsWith('```')) {
        const lines = part.slice(3, -3).split('\n')
        const language = lines[0] || 'text'
        const code = lines.slice(1).join('\n')
        return {
          type: 'codeblock' as const,
          content: code,
          language,
          key: index,
        }
      }

      // Inline code
      if (part.startsWith('`') && part.endsWith('`')) {
        return { type: 'code' as const, content: part.slice(1, -1), key: index }
      }

      // Regular text - parse for other markdown
      const lines = part.split('\n')
      const parsed: ParsedElement[] = []

      lines.forEach((line, lineIndex) => {
        // Headers
        if (line.startsWith('### ')) {
          parsed.push({
            type: 'h3',
            content: line.slice(4),
            key: `${index}-${lineIndex}`,
          })
        } else if (line.startsWith('## ')) {
          parsed.push({
            type: 'h2',
            content: line.slice(3),
            key: `${index}-${lineIndex}`,
          })
        } else if (line.startsWith('# ')) {
          parsed.push({
            type: 'h1',
            content: line.slice(2),
            key: `${index}-${lineIndex}`,
          })
        }
        // Lists
        else if (line.match(/^[*\-+]\s/)) {
          parsed.push({
            type: 'list-item',
            content: line.slice(2),
            key: `${index}-${lineIndex}`,
          })
        }
        // Numbered lists
        else if (line.match(/^\d+\.\s/)) {
          parsed.push({
            type: 'ordered-list-item',
            content: line.replace(/^\d+\.\s/, ''),
            key: `${index}-${lineIndex}`,
          })
        }
        // Blockquotes
        else if (line.startsWith('> ')) {
          parsed.push({
            type: 'blockquote',
            content: line.slice(2),
            key: `${index}-${lineIndex}`,
          })
        }
        // Tables (simple detection)
        else if (line.includes('|') && line.trim().startsWith('|')) {
          parsed.push({
            type: 'table-row',
            content: line,
            key: `${index}-${lineIndex}`,
          })
        }
        // Regular paragraph
        else if (line.trim()) {
          parsed.push({
            type: 'text',
            content: line,
            key: `${index}-${lineIndex}`,
          })
        }
      })

      return parsed
    })
    .flat()
}

type InterpretationType = {
  interpretation: string
}

function Interpretation({ interpretation }: InterpretationType) {
  const parsedContent = useMemo(
    () => parseMarkdown(interpretation),
    [interpretation],
  )

  // Group consecutive list items
  const groupedContent = useMemo((): GroupedElement[] => {
    const result: GroupedElement[] = []
    let currentList: string[] = []
    let currentListType: 'unordered' | 'ordered' | null = null

    parsedContent.forEach((item) => {
      if (item.type === 'list-item') {
        if (currentListType !== 'unordered') {
          if (currentList.length > 0 && currentListType) {
            result.push({ type: currentListType, items: currentList })
          }
          currentList = [item.content]
          currentListType = 'unordered'
        } else {
          currentList.push(item.content)
        }
      } else if (item.type === 'ordered-list-item') {
        if (currentListType !== 'ordered') {
          if (currentList.length > 0 && currentListType) {
            result.push({ type: currentListType, items: currentList })
          }
          currentList = [item.content]
          currentListType = 'ordered'
        } else {
          currentList.push(item.content)
        }
      } else {
        if (currentList.length > 0 && currentListType) {
          result.push({ type: currentListType, items: currentList })
          currentList = []
          currentListType = null
        }
        result.push(item)
      }
    })

    if (currentList.length > 0 && currentListType) {
      result.push({ type: currentListType, items: currentList })
    }

    return result
  }, [parsedContent])

  // Function to render bold and italic text
  const renderFormattedText = (text: string) => {
    // Handle bold and italic
    const parts = text.split(/(\*\*[^*]+\*\*|\*[^*]+\*|__[^_]+__|_[^_]+_)/g)

    return parts.map((part, i) => {
      if (part.startsWith('**') && part.endsWith('**')) {
        return (
          <Text key={i} component="span" fw={700} className={'mx-[0.5ch]'}>
            {part.slice(2, -2)}
          </Text>
        )
      }
      if (part.startsWith('__') && part.endsWith('__')) {
        return (
          <Text key={i} component="span" fw={700} className={'mx-[0.5ch]'}>
            {part.slice(2, -2)}
          </Text>
        )
      }
      if (
        part.startsWith('*') &&
        part.endsWith('*') &&
        !part.startsWith('**')
      ) {
        return (
          <Text key={i} component="span" fs="italic" className={'mx-[0.5ch]'}>
            {part.slice(1, -1)}
          </Text>
        )
      }
      if (
        part.startsWith('_') &&
        part.endsWith('_') &&
        !part.startsWith('__')
      ) {
        return (
          <Text key={i} component="span" fs="italic" className={'mx-[0.5ch]'}>
            {part.slice(1, -1)}
          </Text>
        )
      }
      // Handle highlights
      if (part.includes('==')) {
        const highlighted = part.split(/==[^=]+==/g)
        return highlighted.map((h, hi) => {
          if (h.startsWith('==') && h.endsWith('==')) {
            return (
              <Mark key={`${i}-${hi}`} color="yellow">
                {h.slice(2, -2)}
              </Mark>
            )
          }
          return h
        })
      }
      return part
    })
  }

  const renderContent = () => {
    return groupedContent.map((item) => {
      switch (item.type) {
        case 'h1':
          return (
            <Title
              key={item.key}
              order={2}
              className="mt-6! mb-3! text-2xl font-bold"
            >
              {renderFormattedText(item.content)}
            </Title>
          )
        case 'h2':
          return (
            <Title
              key={item.key}
              order={3}
              className="mt-5 mb-2 text-xl font-semibold"
            >
              {renderFormattedText(item.content)}
            </Title>
          )
        case 'h3':
          return (
            <Title
              key={item.key}
              order={4}
              className="mt-4 mb-2 text-lg font-medium"
            >
              {renderFormattedText(item.content)}
            </Title>
          )
        case 'text':
          return (
            <Text key={item.key} size="md" className="mb-3! leading-relaxed">
              {renderFormattedText(item.content)}
            </Text>
          )
        case 'code':
          return (
            <Code key={item.key} className="px-2 py-1 mx-1">
              {item.content}
            </Code>
          )
        case 'codeblock':
          return (
            <Paper
              key={item.key}
              radius="md"
              withBorder
              className="mb-4! overflow-hidden"
            >
              <Group
                justify="space-between"
                className="px-3 py-2 bg-gray-100 dark:bg-gray-800 border-b"
              >
                <Group gap="xs">
                  <IconTerminal2 size={16} className="text-gray-600" />
                  <Text size="xs" c="dimmed" tt="uppercase" fw={600}>
                    {item.language}
                  </Text>
                </Group>
                <CopyButton value={item.content}>
                  {({ copied, copy }) => (
                    <Tooltip label={copied ? 'Copied' : 'Copy code'}>
                      <ActionIcon
                        color={copied ? 'teal' : 'gray'}
                        onClick={copy}
                        variant="subtle"
                        size="sm"
                      >
                        {copied ? (
                          <IconCheck size={16} />
                        ) : (
                          <IconCopy size={16} />
                        )}
                      </ActionIcon>
                    </Tooltip>
                  )}
                </CopyButton>
              </Group>
              <ScrollArea className="max-h-96">
                <Code block className="p-4 text-sm">
                  {item.content}
                </Code>
              </ScrollArea>
            </Paper>
          )
        case 'unordered':
          return (
            <List key={item.key} spacing="xs" size="md" className="mb-4!">
              {item.items.map((listItem: string, i: number) => (
                <List.Item key={i}>{renderFormattedText(listItem)}</List.Item>
              ))}
            </List>
          )
        case 'ordered':
          return (
            <List
              key={item.key}
              type="ordered"
              spacing="xs"
              size="md"
              className="mb-4!"
            >
              {item.items.map((listItem: string, i: number) => (
                <List.Item key={i}>{renderFormattedText(listItem)}</List.Item>
              ))}
            </List>
          )
        case 'blockquote':
          return (
            <Blockquote
              key={item.key}
              color="blue"
              icon={<IconQuote size={20} />}
              className="my-4"
            >
              {renderFormattedText(item.content)}
            </Blockquote>
          )
        case 'table-row': {
          // Simple table rendering (you might want to enhance this)
          const cells = item.content.split('|').filter((c: string) => c.trim())
          return (
            <Table.Tr key={item.key}>
              {cells.map((cell: string, i: number) => (
                <Table.Td key={i}>{cell.trim()}</Table.Td>
              ))}
            </Table.Tr>
          )
        }
        default:
          return null
      }
    })
  }

  return (
    <Stack gap="lg" className="w-full">
      {/* Header */}
      <Group justify="space-between" className="mb-2">
        <Group gap="xs">
          <ThemeIcon
            size="lg"
            radius="md"
            variant="gradient"
            gradient={{ from: 'indigo', to: 'cyan' }}
          >
            <IconBrain size={20} />
          </ThemeIcon>
          <Text size="xl" fw={700}>
            AI suggestion
          </Text>
        </Group>
      </Group>

      {/* Main Content Card */}
      <Card
        shadow="sm"
        padding="xl"
        radius="md"
        withBorder
        className="relative overflow-hidden"
      >
        {/* Decorative gradient background */}
        <div className="absolute top-0 right-0 w-64 h-64 bg-gradient-to-br from-blue-500/5 to-purple-500/5 rounded-full blur-3xl" />

        {/* Main Answer Content */}
        <Group gap="xs" className="mb-4!">
          <IconFileText size={20} className="text-gray-600" />
          <Text size="sm" c="dimmed" tt="uppercase" fw={600}>
            Response
          </Text>
          <CopyButton value={interpretation}>
            {({ copied, copy }) => (
              <Tooltip label={copied ? 'Copied!' : 'Copy entire answer'}>
                <ActionIcon
                  color={copied ? 'teal' : 'gray'}
                  onClick={copy}
                  variant="subtle"
                  size="sm"
                  className="ml-auto"
                >
                  {copied ? <IconCheck size={16} /> : <IconCopy size={16} />}
                </ActionIcon>
              </Tooltip>
            )}
          </CopyButton>
        </Group>

        <Divider className="mb-4!" />

        {/* Rendered Content */}
        <Box className="prose prose-lg max-w-none">{renderContent()}</Box>

        {/* Footer Info */}
        <Group justify="space-between" className="mt-6!">
          <Group gap="xs">
            <IconInfoCircle size={16} className="text-gray-500" />
            <Text size="xs" c="dimmed">
              Response length: {interpretation.length} characters
            </Text>
          </Group>
          <Text size="xs" c="dimmed">
            Processed at {new Date().toLocaleTimeString()}
          </Text>
        </Group>
      </Card>
    </Stack>
  )
}

export default React.memo(Interpretation)
