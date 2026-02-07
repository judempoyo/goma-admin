import type { HttpMethod } from "~/types/http"

export async function useApiFetch(
    url: string,
    options: {
        method?: HttpMethod
        [key: string]: any
    } = {}
) {
    const { $api } = useNuxtApp()

    return await $api(url, {
        method: options.method ?? "GET",
        ...options,
    })
}