import { useState, useEffect, useCallback } from 'react'
import { Link2, Unlink, Loader2, UserRound } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Separator } from '@/components/ui/separator'
import { ConfirmDialog } from '@/components/confirm-dialog'
import { getUser, adminClearUserBinding } from '../../api'
import type { User } from '../../types'

interface Props {
  open: boolean
  onOpenChange: (open: boolean) => void
  userId: number | null
  onUnbindSuccess?: () => void
}

interface BindingItem {
  key: string
  label: string
  icon: React.ReactNode
  value: string
  isBound: boolean
}

const BUILTIN_BINDINGS: ReadonlyArray<{
  key: string
  field: string
  label: string
  icon: React.ReactNode
}> = [
  {
    key: 'sparkloc',
    field: 'sparkloc_id',
    label: 'Sparkloc',
    icon: <UserRound className='h-4 w-4' />,
  },
]

export function UserBindingDialog(props: Props) {
  const { t } = useTranslation()
  const [user, setUser] = useState<User | null>(null)
  const [loading, setLoading] = useState(false)
  const [unbindTarget, setUnbindTarget] = useState<BindingItem | null>(null)
  const [unbinding, setUnbinding] = useState(false)

  const fetchData = useCallback(async () => {
    if (!props.userId) return
    setLoading(true)
    try {
      const userRes = await getUser(props.userId)
      if (userRes.success && userRes.data) {
        setUser(userRes.data)
      }
    } catch {
      toast.error(t('Failed to load'))
    } finally {
      setLoading(false)
    }
  }, [props.userId, t])

  useEffect(() => {
    if (props.open && props.userId) {
      fetchData()
    } else {
      setUser(null)
    }
  }, [props.open, props.userId, fetchData])

  const allBindings = (() => {
    const items: BindingItem[] = []

    for (const field of BUILTIN_BINDINGS) {
      const value = user
        ? String((user as Record<string, unknown>)[field.field] || '')
        : ''
      const isBound = !!value

      items.push({
        key: field.key,
        label: field.label,
        icon: field.icon,
        value: isBound ? value : '',
        isBound,
      })
    }

    return items
  })()

  const boundCount = allBindings.filter((b) => b.isBound).length

  const handleUnbind = async () => {
    if (!unbindTarget || !props.userId) return
    setUnbinding(true)
    try {
      const res = await adminClearUserBinding(props.userId, unbindTarget.key)
      if (res?.success) {
        toast.success(
          t('Unbound {{provider}}', { provider: unbindTarget.label })
        )
        await fetchData()
        props.onUnbindSuccess?.()
      } else {
        toast.error(res?.message || t('Unbind failed'))
      }
    } catch {
      toast.error(t('Unbind failed'))
    } finally {
      setUnbinding(false)
      setUnbindTarget(null)
    }
  }

  return (
    <>
      <Dialog open={props.open} onOpenChange={props.onOpenChange}>
        <DialogContent className='sm:max-w-lg'>
          <DialogHeader>
            <DialogTitle className='flex items-center gap-2'>
              <Link2 className='h-5 w-5' />
              {t('Account Binding Management')}
            </DialogTitle>
            <DialogDescription className='sr-only'>
              {t('Manage account bindings for this user')}
            </DialogDescription>
          </DialogHeader>

          {loading ? (
            <div className='flex items-center justify-center py-8'>
              <Loader2 className='text-muted-foreground h-6 w-6 animate-spin' />
            </div>
          ) : (
            <div className='space-y-3'>
              <div className='flex items-center justify-between'>
                {user && (
                  <p className='text-muted-foreground text-sm'>
                    {user.username} (ID: {user.id})
                  </p>
                )}
              </div>

              <Separator />

              <div>
                {allBindings.length === 0 ? (
                  <p className='text-muted-foreground py-4 text-center text-sm'>
                    {t('No providers available')}
                  </p>
                ) : (
                  <div className='grid grid-cols-1 gap-2 pr-3 lg:grid-cols-2'>
                    {allBindings.map((binding) => (
                      <div
                        key={binding.key}
                        className={`flex items-center justify-between rounded-md border px-3 py-2.5 ${
                          !binding.isBound ? 'opacity-50' : ''
                        }`}
                      >
                        <div className='flex min-w-0 items-center gap-2.5'>
                          <div className='text-muted-foreground shrink-0'>
                            {binding.icon}
                          </div>
                          <div className='min-w-0'>
                            <div className='flex items-center gap-1.5'>
                              <span className='text-sm font-medium'>
                                {binding.label}
                              </span>
                            </div>
                            <p className='text-muted-foreground max-w-[140px] truncate text-xs'>
                              {binding.isBound ? binding.value : t('Not bound')}
                            </p>
                          </div>
                        </div>
                        {binding.isBound && (
                          <Button
                            variant='ghost'
                            size='sm'
                            className='text-destructive hover:text-destructive h-7 w-7 shrink-0 p-0'
                            onClick={() => setUnbindTarget(binding)}
                          >
                            <Unlink className='h-3.5 w-3.5' />
                          </Button>
                        )}
                      </div>
                    ))}
                  </div>
                )}
              </div>

              <p className='text-muted-foreground text-xs'>
                {t('Bound')}: {boundCount} / {allBindings.length}
              </p>
            </div>
          )}
        </DialogContent>
      </Dialog>

      <ConfirmDialog
        open={!!unbindTarget}
        onOpenChange={(open) => !open && setUnbindTarget(null)}
        title={t('Confirm Unbind')}
        desc={t(
          'Are you sure you want to unbind {{provider}} for this user? The user will no longer be able to log in via this method.',
          {
            provider: unbindTarget?.label || '',
          }
        )}
        confirmText={t('Confirm Unbind')}
        destructive
        handleConfirm={handleUnbind}
        isLoading={unbinding}
      />
    </>
  )
}
