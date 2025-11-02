// #region /search
export type SearchRequest = {
  q: string
  teams: string[]
  levels: string[]
  locations: string[]
  limit: number
}

export type SearchResult = {
  score: number
  user: User
  description: string
}

// #endregion

// #region User

export type User = {
  _id: string
  email: string
  password: string
  name: string
  providers: UserProvider[]
  verified: boolean
  status: UserStatus
  semantic: UserSemantic
  contact: UserContact
  avatarId: string
  organizationId: string
  organizationRole: OrganizationRole
}

export type UserProvider = 'email' | 'google'

export const UserProviders = {
  Email: 'email' as UserProvider,
  Google: 'google' as UserProvider,
}

export type UserStatus = 'ACTIVE' | 'DELETED'

export const UserStatuses = {
  ACTIVE: 'ACTIVE' as UserStatus,
  DELETED: 'DELETED' as UserStatus,
}

export type UserSemantic = {
  description: string
  team: string
  level: string
  location: string
}

export type UserContact = {
  slack: string
  telephone: string
  email: string
  telegram: string
  whatsapp: string
}

export const mockUser: User = {
  _id: '123',
  email: 'msk.vitaly@gmail.com',
  password: 'hashed_password',
  name: 'Vitalii Popov',
  providers: ['email', 'google'],
  verified: true,
  status: 'ACTIVE',
  semantic: {
    description: 'Full Stack developer',
    team: 'team_6',
    level: 'level_3',
    location: 'Chemnitz',
  },
  contact: {
    slack: 'vitaliipopov',
    telephone: '+491745624691',
    email: 'msk.vitaly@gmail.com',
    telegram: '@mskVitalii',
    whatsapp: '+491745624691',
  },
  avatarId: 'avatar_001',
  organizationId: 'org_001',
  organizationRole: 'OWNER',
}

// #endregion

// #region Organization
export type OrganizationRole = 'OWNER' | 'ADMIN' | 'USER'

export const OrganizationRoles = {
  OWNER: 'OWNER' as OrganizationRole,
  ADMIN: 'ADMIN' as OrganizationRole,
  USER: 'USER' as OrganizationRole,
}

export type OrganizationPlanType = 'FREE' | 'BUSINESS'

export const OrganizationPlans = {
  FREE: 'FREE' as OrganizationPlanType,
  BUSINESS: 'BUSINESS' as OrganizationPlanType,
}

export type Organization = {
  id: string
  title: string
  semantic: OrganizationSemantic
  plan: OrganizationPlanType
}

export type OrganizationSemantic = {
  levels: Level[]
  teams: Team[]
  locations: Location[]
}

export type Level = {
  id?: string
  name: string
  description: string
}

export type Team = {
  id?: string
  name: string
  description: string
}

export type Location = {
  id?: string
  name: string
}

export const mockLocations = [
  { name: 'Chemnitz' },
  { name: 'Berlin' },
  { name: 'Dresden' },
  { name: 'Prague' },
  { name: 'Amsterdam' },
]
export const mockTeams = [
  {
    id: 'team_1',
    name: 'Intranet Experience',
    description:
      'Responsible for the core Staffbase intranet product — delivering employee communications, content management, and engagement features used across enterprise clients.',
  },
  {
    id: 'team_2',
    name: 'Microsoft Integration',
    description:
      'Develops and maintains integrations between Staffbase and Microsoft 365, including Teams, SharePoint, and Outlook add-ins to unify digital workplace tools.',
  },
  {
    id: 'team_3',
    name: 'Mobile Platform',
    description:
      'Owns the Staffbase mobile app architecture, release pipeline, and feature development for both iOS and Android clients.',
  },
  {
    id: 'team_4',
    name: 'Analytics & Insights',
    description:
      'Builds dashboards, analytics pipelines, and data models that help organizations measure engagement and communication impact.',
  },
  {
    id: 'team_5',
    name: 'Customer Success Engineering',
    description:
      'Works closely with enterprise customers to ensure smooth onboarding, integrations, and custom feature delivery within the Staffbase ecosystem.',
  },
  {
    id: 'team_6',
    name: 'AI',
    description: 'Architecture. AI. New scallable products',
  },
]

export const mockLevel = [
  {
    id: 'level_1',
    name: 'Intern',
    description: 'Working student, minor tasks',
  },
  { id: 'level_2', name: 'Junior', description: 'Entry-level developer' },
  { id: 'level_3', name: 'Middle', description: 'Independent contributor' },
  {
    id: 'level_4',
    name: 'Senior',
    description: 'Expert developer, mentoring others',
  },
  {
    id: 'level_5',
    name: 'Staff',
    description: 'Tech leadership and architecture',
  },
]

export const mockOrganization: Organization = {
  id: 'org_1',
  title: 'Staffbase',
  plan: OrganizationPlans.BUSINESS,
  semantic: {
    levels: mockLevel,
    teams: mockTeams,
    locations: mockLocations,
  },
}

// #endregion
