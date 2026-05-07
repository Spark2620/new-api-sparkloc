import { createFileRoute, redirect } from '@tanstack/react-router'

export const Route = createFileRoute('/(auth)/user/reset')({
  beforeLoad: () => {
    throw redirect({ to: '/sign-in' })
  },
})
