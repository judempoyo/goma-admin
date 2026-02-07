<script setup lang="ts">
import { zodResolver } from '@primevue/forms/resolvers/zod';
import { z } from 'zod';
definePageMeta({
  layout: 'guest',
  middleware: 'guest',
});

const toast = useToast();
const loading = ref(false);

const route = useRoute();
const router = useRouter();

const initialValues = reactive({
  password: '',
  confirm_password: '',
});

const resolver = ref(
  zodResolver(
    z
      .object({
        password: z.string().min(8, { message: 'Password must be at least 8 characters.' }),
        confirm_password: z.string().min(1, { message: 'Confirmation is required.' }),
      })
      .refine((data) => data.password === data.confirm_password, {
        message: "Passwords don't match",
        path: ['confirm_password'],
      })
  )
);

onMounted(() => {
  const token = Array.isArray(route.query.token) ? route.query.token[0] : route.query.token;
  if (!token) {
    toast.add({
      severity: 'error',
      summary: 'Token Missing',
      detail: 'Reset link is missing.',
      life: 4000,
    });
    //router.push('/auth/login');
  }
});

const handleSubmit = async ({ valid, values }:any) => {
  if (!valid) return;
  loading.value = true;
  try {
    await new Promise((resolve) => setTimeout(resolve, 1000));

    toast.add({
      severity: 'success',
      summary: 'Password Reset',
      detail: 'You can now log in with your new password.',
      life: 4000,
    });
    router.push('/auth/login');
  } catch (err: any) {
    toast.add({
      severity: 'error',
      summary: 'Error',
      detail: err.message || 'An error occurred.',
      life: 4000,
    });
  } finally {
    loading.value = false;
  }
};
</script>

<template>
    <transition name="fade-scale" appear>
      <div class="w-full max-w-md p-6 sm:p-8 rounded-2xl bg-surface-0 dark:bg-surface-900 shadow-md flex flex-col gap-6">

        <div class="flex flex-col items-center gap-2 text-center mb-4">
          <NuxtLink to="/">
          <img src="/favicon.ico" class="h-16 w-16 sm:h-20 sm:w-20" alt="Logo" />
        </NuxtLink>
          <h2 class="text-lg font-semibold text-surface-900 dark:text-surface-0">Set a New Password</h2>
          <p class="text-sm text-surface-700 dark:text-surface-200 px-2">
            Choose a strong password (minimum 8 characters). Confirm your password below.
          </p>
        </div>

        <Form v-slot="$form" :initialValues :resolver="resolver" @submit="handleSubmit" class="flex flex-col gap-4 w-full">

          <div class="flex flex-col gap-1 w-full">
            <label for="password" class="text-sm font-medium text-surface-900 dark:text-surface-0">Password</label>
            <Password
              id="password"
              name="password"
              placeholder="Enter a strong password"
              toggleMask
              :feedback="false"
              fluid
            />
            <Message v-if="$form.password?.invalid" severity="error" size="small" variant="simple">
              {{ $form.password.error?.message }}
            </Message>
          </div>

          <div class="flex flex-col gap-1 w-full">
            <label for="confirm_password" class="text-sm font-medium text-surface-900 dark:text-surface-0">Confirm Password</label>
            <Password
              id="confirm_password"
              name="confirm_password"
              placeholder="Confirm your password"
              toggleMask
              :feedback="false"
              fluid
            />
            <Message v-if="$form.confirm_password?.invalid" severity="error" size="small" variant="simple">
              {{ $form.confirm_password.error?.message }}
            </Message>
          </div>

          <Button
            type="submit"
            severity="primary"
            label="Reset Password"
            class="w-full rounded-xl py-3"
            :loading="loading"
          />

        </Form>

        <a href="/auth/login" class="text-sm text-primary group transition-all mt-6 flex items-center justify-center gap-1">
          <i class="pi pi-arrow-left group-hover:-translate-x-1"></i> 
          <span class="group-hover:underline underline-offset-4">
          Back to Login
          </span> 
        </a>

      </div>
    </transition>
</template>

<style scoped>
</style>
