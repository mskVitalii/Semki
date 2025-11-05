import type { SearchRequest } from '@/common/types'
import { useOrganizationStore } from '@/stores/organizationStore'
import {
  Button,
  Group,
  Loader,
  MultiSelect,
  Paper,
  Stack,
  TextInput,
  Tooltip,
} from '@mantine/core'
import { useForm } from '@mantine/form'
import { IconPlayerStop, IconSend } from '@tabler/icons-react'
import { useEffect, useRef } from 'react'
import { useParams } from 'react-router-dom'

type SearchFormProps = {
  onSearch: (
    query: string,
    filters?: Omit<SearchRequest, 'q' | 'limit'>,
  ) => void
  onCancel: () => void
  isLoading: boolean
  req?: SearchRequest
}

function SearchForm({ req, onSearch, onCancel, isLoading }: SearchFormProps) {
  const { chatId } = useParams<{ chatId?: string }>()
  const organization = useOrganizationStore((s) => s.organization)
  const qPlaceholder = useRef(getRandomPlaceholder())

  const form = useForm<SearchRequest>({
    initialValues: {
      q: req?.q ?? '',
      teams: req?.teams ?? [],
      levels: req?.levels ?? [],
      locations: req?.locations ?? [],
      limit: req?.limit ?? 10,
    },
  })

  useEffect(() => {
    form.setValues({
      q: '',
      teams: [],
      levels: [],
      locations: [],
      limit: 10,
    })
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [chatId])

  const handleSubmit = (values: SearchRequest) => {
    if (!values.q.trim()) return
    onSearch(values.q.trim(), {
      teams: values.teams,
      levels: values.levels,
      locations: values.locations,
    })
  }
  if (!organization)
    return (
      <div className="flex-1 flex items-center justify-center h-full">
        <Loader color="green" />
      </div>
    )

  return (
    <Paper p="md" radius="md" withBorder className="bg-gray-50">
      <form onSubmit={form.onSubmit(handleSubmit)}>
        <Stack gap="md">
          <TextInput
            label="Who should I find?"
            placeholder={qPlaceholder.current}
            rightSection={isLoading && <Loader c="green" size="xs" />}
            variant="unstyled"
            classNames={{
              label: 'text-2xl!',
            }}
            size="lg"
            {...form.getInputProps('q')}
          />

          <Group align="flex-start" gap="md">
            <Group grow align="flex-start">
              <MultiSelect
                w={'100%'}
                placeholder="Select teams"
                data={organization.semantic.teams.map((t) => ({
                  value: t.id ?? t.name,
                  label: t.name,
                }))}
                searchable
                {...form.getInputProps('teams')}
              />
            </Group>
            <Group grow align="flex-start">
              <MultiSelect
                placeholder="Select levels"
                data={organization.semantic.levels.map((l) => ({
                  value: l.id ?? l.name,
                  label: l.name,
                }))}
                searchable
                {...form.getInputProps('levels')}
              />
              <MultiSelect
                placeholder="Select locations"
                data={organization.semantic.locations.map((l) => ({
                  value: l.name,
                  label: l.name,
                }))}
                searchable
                {...form.getInputProps('locations')}
              />
            </Group>
          </Group>

          <Group justify="flex-end">
            {isLoading ? (
              <Tooltip label="Stop streaming">
                <Button
                  onClick={onCancel}
                  color="red"
                  leftSection={<IconPlayerStop size={18} />}
                  size="md"
                  w="100%"
                >
                  Stop
                </Button>
              </Tooltip>
            ) : (
              <Tooltip label="Send question">
                <Button
                  w="100%"
                  type="submit"
                  disabled={!form.values.q.trim()}
                  leftSection={<IconSend size={18} />}
                  size="md"
                  color="green"
                >
                  Search
                </Button>
              </Tooltip>
            )}
          </Group>
        </Stack>
      </form>
    </Paper>
  )
}

const placeholders = [
  'Find me a partner to eat pasta on lunch! 🍝(˶ᐢ ᵕ ᐢ˶)',
  'Find the right person to own it! 💼( ^_^)',
  'Find the White Rabbit >>> ૮꒰ ˶• ༝ •˶꒱ა ♡',
  'Find my Morty. Wubba Lubba Dub Dub! (☞0_0)☞',
  'Wake up, Samurai. I have contacts to talk 🗡️(⌐■_■)',
]

function getRandomPlaceholder() {
  const index = Math.floor(Math.random() * placeholders.length)
  return placeholders[index]
}

export default SearchForm
