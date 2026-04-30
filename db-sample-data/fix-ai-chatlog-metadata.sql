-- Fix ai_chatlog schema drift in the Manager DB.
-- Symptom:
--   ERROR: column "metadata" of relation "ai_chatlog" does not exist (SQLSTATE 42703)
--
-- This file is intentionally non-destructive:
--   - It creates ai_chatlog only when the table is missing.
--   - It adds metadata only when the column is missing.
--   - It does not truncate, drop, or reinitialize any data.

CREATE TABLE IF NOT EXISTS public.ai_chatlog (
    id bigserial PRIMARY KEY,
    session_id varchar(100) NOT NULL,
    user_id varchar(100),
    provider varchar(50) NOT NULL DEFAULT 'twcc',
    model varchar(100),
    question text NOT NULL,
    answer text,
    tool_used boolean DEFAULT false,
    tools jsonb,
    input_tokens integer DEFAULT 0,
    output_tokens integer DEFAULT 0,
    total_tokens integer DEFAULT 0,
    latency_ms integer,
    status varchar(30) NOT NULL DEFAULT 'success',
    error_code varchar(100),
    error_message text,
    metadata jsonb DEFAULT '{}'::jsonb,
    ip_address varchar(45) NOT NULL,
    created_at timestamp with time zone NOT NULL DEFAULT now()
);

ALTER TABLE public.ai_chatlog
    ADD COLUMN IF NOT EXISTS metadata jsonb DEFAULT '{}'::jsonb;

UPDATE public.ai_chatlog
SET metadata = '{}'::jsonb
WHERE metadata IS NULL;

CREATE INDEX IF NOT EXISTS idx_ai_chatlog_session
    ON public.ai_chatlog (session_id);

CREATE INDEX IF NOT EXISTS idx_ai_chatlog_user
    ON public.ai_chatlog (user_id);
