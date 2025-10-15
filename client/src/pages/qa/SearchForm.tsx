import { Button, Group, Loader, Paper, TextInput, Tooltip } from '@mantine/core'
import { IconPlayerStop, IconSend } from '@tabler/icons-react'
import { useState } from 'react'

// TODO: filters

type SearchFormProps = {
  onSearch: (query: string) => void
  onCancel: () => void
  isLoading: boolean
}

function SearchForm({ onSearch, onCancel, isLoading }: SearchFormProps) {
  const [question, setQuestion] = useState<string>()
  const handleKeyPress = (
    event: React.KeyboardEvent<HTMLInputElement>,
  ): void => {
    if (event.key === 'Enter' && !event.shiftKey && question?.trim()) {
      event.preventDefault()
      onSearch(question.trim())
    }
  }

  return (
    <Paper p="md" radius="md" withBorder className="bg-gray-50">
      <Group align="flex-end">
        <TextInput
          className="flex-1"
          label="Which people to find?"
          placeholder={getRandomPlaceholder()}
          rightSectionWidth={130}
          rightSection={<>{isLoading && <Loader size="xs" />}</>}
          mt="md"
          variant="unstyled"
          value={question}
          onChange={(e) => setQuestion(e.currentTarget.value)}
          onKeyDown={handleKeyPress}
          disabled={isLoading}
          size="md"
        />
        {isLoading ? (
          <Tooltip label="Stop streaming">
            <Button
              onClick={onCancel}
              color="red"
              leftSection={<IconPlayerStop size={18} />}
              size="md"
            >
              Stop
            </Button>
          </Tooltip>
        ) : (
          <Tooltip label="Send question">
            <Button
              onClick={() => question?.trim() && onSearch(question)}
              disabled={!question?.trim()}
              leftSection={<IconSend size={18} />}
              size="md"
            >
              Send
            </Button>
          </Tooltip>
        )}
      </Group>
    </Paper>
  )
}

//#region Placeholders
const placeholders = [
  'Find me a partner to eat pasta on lunch! 🍝(˶ᐢ ᵕ ᐢ˶)',
  'Find the right person to own it! 💼( ^_^)',
  'Find the White Rabbit >>> ૮꒰ ˶• ༝ •˶꒱ა ♡',
  'Find your Morty. Wubba Lubba Dub Dub! (☞0_0)☞',
  'Wake up, Samurai. You have contacts to talk 🗡️(⌐■_■)',
  'Find friends. Say hello to your little friend! 🔫(｀ω´)',
  'Find John Connor 🤖( •_•)>⌐■-■',
  "Find who's breathtaking 💫(˶ˊᵕˋ˵)",
  'If the cake is a lie, find the baker 🍰(´･ω･`)',
  'Find the detonator 🃏(¬‿¬)',
  'Search the infinity and beyond! 🚀(•̀ᴗ•́)و',
  'Find Gandalf before it’s too late 🧙‍♂️(╯°□°）╯︵ ┻━┻',
]

function getRandomPlaceholder() {
  const index = Math.floor(Math.random() * placeholders.length)
  return placeholders[index]
}
//#endregion

export default SearchForm
