import { createFileRoute, redirect } from '@tanstack/react-router'

export const Route = createFileRoute('/(auth)/oauth')({
  beforeLoad: () => {
    throw redirect({ to: '/sign-in' })
  },
})
