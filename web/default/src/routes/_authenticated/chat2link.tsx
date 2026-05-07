import { createFileRoute, redirect } from '@tanstack/react-router'

export const Route = createFileRoute('/_authenticated/chat2link')({
  beforeLoad: () => {
    throw redirect({ to: '/dashboard' })
  },
})
