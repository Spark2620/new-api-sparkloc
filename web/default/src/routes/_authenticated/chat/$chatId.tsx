import { createFileRoute, redirect } from '@tanstack/react-router'

export const Route = createFileRoute('/_authenticated/chat/$chatId')({
  beforeLoad: () => {
    throw redirect({ to: '/dashboard' })
  },
})
