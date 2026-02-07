<script setup lang="ts">
import { zodResolver } from '@primevue/forms/resolvers/zod';
import { z } from 'zod';
definePageMeta({
  layout: 'guest',
  middleware: 'guest',
});

const toast = useToast();
const loading = ref(false);
const emailSent = ref(false);
const sentTo = ref('');

const initialValues = reactive({ email: '' });

const resolver = ref(zodResolver(
    z.object({
        email: z.email({ message: 'Invalid email address.' })

    })
));


const handleSubmit = async ({ valid, values }:any) => {
  if (!valid) return;
  loading.value = true;
  try {
    await new Promise((resolve) => setTimeout(resolve, 1000));
    sentTo.value = values.email;
    emailSent.value = true;

    toast.add({
      severity: 'success',
      summary: 'Link Sent',
      detail: `Instructions sent to ${values.email}`,
      life: 4000,
    });
  } finally {
    loading.value = false;
  }
};

const handleResend = async () => {
  loading.value = true;
  try {
    await new Promise((resolve) => setTimeout(resolve, 1000));
    toast.add({
      severity: 'success',
      summary: 'Resent',
      detail: `Link resent to ${sentTo.value}`,
      life: 3000,
    });
  } finally {
    loading.value = false;
  }
};
</script>

<template>
  <transition name="fade-scale" appear>
    <div class="w-full max-w-md p-6 sm:p-8 rounded-2xl bg-surface-0 dark:bg-surface-900 shadow-md flex flex-col gap-6">

      <div v-if="emailSent" class="flex flex-col items-center gap-4 text-center">
        <div class="w-16 h-16 bg-surface-100 dark:bg-surface-700 rounded-full flex items-center justify-center">
          <i class="pi pi-check text-primary text-2xl"></i>
        </div>

        <h2 class="text-lg font-semibold text-surface-900 dark:text-surface-0">Check Your Inbox</h2>
        <p class="text-surface-700 dark:text-surface-200">
          A password reset link has been sent to <br>
          <span class="font-medium text-primary">{{ sentTo }}</span>.
        </p>

        <p class="text-sm text-surface-600 dark:text-surface-300">
          Didn't receive the email?
          <button
            class="text-primary font-medium hover:underline ml-1"
            @click="handleResend"
            :disabled="loading"
          >
            Resend
          </button>
        </p>

        <a href="/auth/login" class="text-sm text-primary hover:underline mt-4 flex items-center gap-1">
          <i class="pi pi-arrow-left"></i> Back to Login
        </a>
      </div>

      <div v-else>
        <div class="flex flex-col items-center gap-2 text-center mb-4">
        <NuxtLink to="/">
          <img src="/favicon.ico" class="h-16 w-16 sm:h-20 sm:w-20" alt="Logo" />
        </NuxtLink>
          <h2 class="text-lg font-semibold text-surface-900 dark:text-surface-0">Forgot Password?</h2>
          <p class="text-sm text-surface-700 dark:text-surface-200 px-2">
            Enter your email address and we’ll send you a link to reset your password.
          </p>
        </div>

        <Form v-slot="$form" :initialValues :resolver @submit="handleSubmit" class="flex flex-col gap-4 w-full">
          <div class="flex flex-col gap-1 w-full">
            <label for="email" class="text-sm font-medium text-surface-900 dark:text-surface-0">Email</label>
            <InputText id="email" name="email" type="email" placeholder="your@email.com" fluid />
            <Message v-if="$form.email?.invalid" severity="error" size="small" variant="simple">
              {{ $form.email.error?.message }}
            </Message>
          </div>

          <Button
            type="submit"
            severity="primary"
            label="Send Reset Link"
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

    </div>
  </transition>
</template>

<style scoped>

</style>
