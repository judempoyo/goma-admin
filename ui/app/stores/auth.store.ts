import type { AuthCredentials, User } from "~/types"

export const useAuthStore = defineStore(`auth`, () => {
    const user = ref<User | null>(null)
    const isLoggedIn = computed(() => !!user.value)
    const token = ref<string | null>(null)
    const refreshToken = ref<string | null>(null)

    function setSession(data: any) {
        if (data.accessToken) {
            token.value = data.accessToken
            localStorage.setItem(`token`, data.accessToken)
        }

        if (data.refreshToken) {
            refreshToken.value = data.refreshToken
            localStorage.setItem(`refreshToken`, data.refreshToken)
        }
    }

    function clearSession() {
        token.value = null
        refreshToken.value = null
        user.value = null

        localStorage.removeItem(`token`)
        localStorage.removeItem(`refreshToken`)
    }


    async function fetchUser() {
        const res:any = await useApiFetch(`/auth/user`)
        user.value = res.data.value
    }

    async function login(credentials: AuthCredentials) {
        const res = await useApiFetch(`/auth/login`, {
            method:"POST",
            body: credentials
        })

        setSession(res)
        await fetchUser()
    }
    async function refreshSession() {

        const res = await useApiFetch(`/auth/refreshToken`, {
            method:"POST",
            body: { refreshToken: refreshToken.value }
        })

        setSession(res)
    }

    async function logout() {
        await useApiFetch(`/auth/logout`, {
            method:`POST`
        })
        clearSession()
        await navigateTo(`/auth/login`)
    }

    return {
        user,
        isLoggedIn,
        token,
        refreshToken,
        login,
        logout,
        fetchUser,
        refreshSession,
        clearSession
    }
})
