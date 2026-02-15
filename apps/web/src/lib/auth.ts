import { z } from "zod/v4";

export const loginSchema = z.object({
  email: z.email("올바른 이메일을 입력해주세요."),
  password: z.string().min(8, "비밀번호는 최소 8자 이상이어야 합니다."),
});

export const signupSchema = z
  .object({
    email: z.email("올바른 이메일을 입력해주세요."),
    password: z
      .string()
      .min(8, "비밀번호는 최소 8자 이상이어야 합니다.")
      .regex(
        /^(?=.*[a-zA-Z])(?=.*\d)/,
        "영문과 숫자를 포함해야 합니다."
      ),
    passwordConfirm: z.string(),
    nickname: z.string().max(50, "닉네임은 50자 이내로 입력해주세요.").optional(),
    agreeToTerms: z.boolean().refine((val) => val === true, {
      message: "이용약관에 동의해주세요.",
    }),
  })
  .refine((data) => data.password === data.passwordConfirm, {
    message: "비밀번호가 일치하지 않습니다.",
    path: ["passwordConfirm"],
  });

export type LoginFormValues = z.infer<typeof loginSchema>;
export type SignupFormValues = z.infer<typeof signupSchema>;
