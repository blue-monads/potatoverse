import { AxiosResponse } from "axios"
import { useEffect, useState } from "react"


type AxiosApi<T> = () => Promise<AxiosResponse<T, any>>

interface PropsType<T> {
    loader: AxiosApi<T>
    ready: boolean
    dependencies?: any[]
}


const useSimpleDataLoader = <T>(props: PropsType<T>) => {
    const [state, setState] = useState<T | null>(null)
    const [loading, setLoading] = useState<boolean>(false)
    const [error, setError] = useState<string | null>(null)

    const load = async () => {
        if (!props.ready) return

        setLoading(true)
        setError(null)

        try {
            const response = await props.loader()
            setState(response.data)
        } catch (err: any) {
            setError(err.message || "An error occurred")
        } finally {
            setLoading(false)
        }
    }

    useEffect(() => {
        load()
    }, [props.ready, ...(props.dependencies ? props.dependencies : [])])

    return {
        data: state,
        loading,
        error,
        reload: load
    }
}


export default useSimpleDataLoader;