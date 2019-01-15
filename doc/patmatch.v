(* Matching verification (in progress)

   See https://github.com/Comcast/sheens#pattern-matching.

   The start of a verification of that pattern matching algorithm.
   This file is written for the Coq Proof Assistant:
   https://coq.inria.fr/.

   The matching here gives a set (list) of returned bindings, but
   currently the matching doesn't support arrays (as sets).

   Status: A work in progress.

 *)

Require Import String Bool Arith List.

(* Require Import SetoidDec. *)

(* for jsCoq: Comments "pkgs: coq-arith".  Then From Coq Require Import String. *)

(* Require Import Cpdt.CpdtTactics. 
   Set Implicit Arguments.
   Set Asymmetric Patterns. *)

Module patmatch.

  Open Scope string_scope.

  Section experiment.

    Inductive exp : Set :=
    | Atom: string -> exp
    | Nil : exp
    | Assoc: (string*exp*exp) -> exp.

    Fixpoint some (k:string) (P:exp->Prop) (alist:list(string*exp)) : Prop :=
      match alist with
      | nil => False
      | (k',v)::more =>
        if string_dec k k'
        then P v
        else some k P more
      end.

    Fixpoint subexp (p:exp) (m:exp) : Prop :=
      match p with
      | Atom s =>
        match m with
        | Atom s' =>
          if string_dec s s' then True else False
        | _ => False
        end
      | Nil =>
        match m with
        | Nil => True
        | _ => False
        end
      | Assoc (k,v,more) =>
        match m with
        | Assoc (k',v',more') =>
          (if string_dec k k'
           then subexp v v'
           else subexp p more')
          /\
          subexp more m
        | _ => False
        end
      end.

    Compute subexp (Atom "tacos") (Atom "queso").

    Compute (Assoc (("likes",(Atom "tacos"))::nil)).

    Compute let e := (Assoc (("likes",(Atom "tacos"))::nil)) in
            subexp e e.
    

  End experiment.

  (* We define our own association list to help with
     well-foundedness arguments for important functions. *)
  
  Section alists.

    Definition alist (T:Set) := list (string*T).

    (* Our association lists. *)
    Definition new_alist {T:Set} : (alist T) := nil.

    (* Add a pair to an alist. *)
    Fixpoint acons {T:Set} (a: (alist T)) (k:string) (v:T) : (alist T) :=
      match a with
      | nil => (k,v)::nil
      | (k',v')::more =>
        if string_dec k k'
        then acons more k v
        else (k',v')::(acons more k v)
      end.

    (* Get the value for a given key. *)
    Fixpoint assoc {T:Set} (a: (alist T)) (k: string) : option T :=
      match a with
      | nil => None
      | (p,v)::more =>
        if string_dec k p
        then Some v
        else assoc more k
      end.

  End alists.

  Section patterns_and_messages.

    (* Expression experiment *)

    Inductive atom : Set :=
    | AStr : string -> atom
    | AVar : string -> atom.
    
    Check AStr "foo".
    
    Inductive exp {A:Set} : Set :=
    | EAtom : A -> exp
    | EMap : (alist (@exp A)) -> exp.

    Check EAtom (AStr "foo").

    Check EAtom "tacos".

    Check EMap new_alist.

    (* Definition mexp := exp (A:=string). *)

    Definition mexp := @exp string.

    Definition pexp := @exp atom.
    
  End patterns_and_messages.

  Section matching.

    (* Is the first message is a sub-message of the second message? *)
    Fixpoint submsg (p:mexp) (m:mexp) : bool :=
      match p, m with
      | EAtom s, EAtom s' =>
        if string_dec s s' then true else false
      | EMap xs, EMap ys =>
        let fix f xs :=
            match xs with
            | nil => true
            | (k,x)::more =>
              match assoc ys k with
                | None => false
                | Some y =>
                  if submsg x y
                  then f more
                  else false
              end
            end
        in f xs
      | _, _ => false
      end.

    Compute submsg (EAtom "tacos") (EAtom "tacos").

    Compute submsg (EAtom "tacos") (EAtom "queso").
    
    Compute submsg (EMap (acons (new_alist) "likes" (EAtom "tacos"))) (EAtom "chips").

    Fixpoint All (p:(string*mexp)->Prop) (xs:list (string*mexp)) : Prop :=
      match xs with
      | nil => True
      | x::more => p x /\ All p more
      end.

    (* Is the first message is a sub-message of the second message? *)
    Program Fixpoint submsgp (p:mexp) (m:mexp) : Prop :=
      match p, m with
      | EAtom s, EAtom s' =>
        if string_dec s s' then True else False
      | EMap xs, EMap ys =>
        let fix p kv :=
            match kv with
            | (k,v) =>
              match assoc ys k with
              | None => False
              | Some y => submsgp v y
              end
            end
        in All p xs
      | _, _ => False
      end.

    Compute submsgp (EAtom "tacos") (EAtom "tacos").

    Compute submsgp (EAtom "tacos") (EAtom "queso").
    
    Compute submsgp (EMap (acons (new_alist) "likes" (EAtom "tacos"))) (EAtom "chips").

    Lemma submsgp_refl : forall x:mexp, submsgp x x.
    Proof.
      intros.
      induction x.
      {
        simpl.
        destruct (string_dec a a).
        reflexivity.
        contradiction.
      }
      {
        induction a.
        {
          simpl.
          trivial.
        }
        {
          
          induction a.
          simpl.
          destruct (string_dec a a).

          
          induction a0.
          reflexivity.
          simpl.
          
    Definition bindings := alist mexp.

    Definition apply_app {A:Type} (lsts:list (list A)) :=
      fold_right (fun x y => x ++ y) nil lsts.
    
    (* The main function. *)
    Fixpoint patmatch (p:pexp) (bs:bindings) (m:mexp) : list bindings :=
      match p with
      | EAtom a =>
        match a with
        | AStr s =>
          match m with
          | EAtom s' =>
            if string_dec s s' then bs::nil else nil
          | _ => nil
          end
        | AVar v =>
          match assoc bs v with
          | None =>
            (acons bs v m)::nil
          | Some m' =>
            if submsg m' m then bs::nil else nil
          end
        end
      | _ => nil
      end.
        
    Definition apply_app {A:Type} (lsts:list (list A)) :=
      fold_right (fun x y => x ++ y) nil lsts.

                                                   
      | PStr ps, Str ms =>
        if string_dec ps ms then bs::nil else nil
      | Var v, _ =>
        match assoc bs v with
        | None => (acons bs v m)::nil
        | Some x => if submsg x m then bs::nil else nil
        end
      | PMap pm, Map mm =>
        let fix f pm bs :=
            match pm with
            | nil => bs::nil
            | (k,v)::pm' =>
              match assoc mm k with
              | None => nil
              | Some v' =>
                apply_app
                  (map (fun (bs:bindings) =>
                          (f pm' bs))
                       (patmatch v bs v'))
              end
            end
        in f pm bs
      | _, _ => nil
      end.

  End matching.

  (* Just some computations to take a look around. *)
  Section patmatch_tests.

    Compute patmatch (PStr "chips") nil (Str "chips") .
    
    Compute let p := (PMap nil) in
            let m := (Map nil) in
            patmatch p nil m.

    Compute let p := (PMap (acons nil "likes" (PStr "tacos"))) in
            let m := (Map (acons nil "likes" (Str "tacos"))) in
            patmatch p nil m.

    Compute let p := (PMap (acons nil "likes" (Var "x"))) in
            let m := (Map (acons nil "likes" (Str "tacos"))) in
            patmatch p nil m.

    Compute let p := (PMap (acons nil "likes" (Var "x"))) in
            let m := (Map (acons (acons nil "wants" (Str "chips"))
                                 "likes" (Str "tacos"))) in
            patmatch p nil m.

    Compute let p := (PMap (acons (acons nil "wants" (Var "y"))
                                  "likes" (Var "x"))) in
            let m := (Map (acons (acons nil "wants" (Str "chips"))
                                 "likes" (Str "tacos"))) in
            patmatch p nil m.

    Compute let p := (PMap (acons (acons (acons nil "needs" (Var "y"))
                                         "wants" (Var "y"))
                                  "likes" (Var "x"))) in
            let m := (Map (acons (acons (acons nil "needs" (Str "chips"))
                                        "wants" (Str "chips"))
                                 "likes" (Str "tacos"))) in
            patmatch p nil m.

    Compute let p := (PMap (acons (acons (acons nil "needs" (Var "y"))
                                         "wants" (Var "y"))
                                  "likes" (Var "x"))) in
            let m := (Map (acons (acons (acons nil "needs" (Str "queso"))
                                        "wants" (Str "chips"))
                                 "likes" (Str "tacos"))) in
            patmatch p nil m.


  End patmatch_tests.

  Section verification.

    (* assoc after an acons does what you'd expect. *)
    Remark acons_assoc :
      forall m:list (string*msg),
      forall k:string,
      forall v:msg,
        (assoc (acons m k v) k) = Some v.
      intros.
      induction m.
      simpl.
      destruct (string_dec k k).
      reflexivity.
      intuition.
      simpl.
      induction a.
      destruct (string_dec k a).
      assumption.
      simpl.
      destruct (string_dec k a).
      contradiction.
      assumption.
    Defined.

    (* Do the bindings applied to the pattern give a submsg of the
       message? *)
    Fixpoint psubmsg (p:pat) (bs:bindings) (m:msg) : bool :=
      match p, m with
      | PStr ps, Str ms =>
        if string_dec ps ms then true else false
      | Var k, _ =>
        match assoc bs k with
        | None => false
        | Some v => submsg v m
        end
      | PMap pm, Map mm =>
        let fix f pm :=
            match pm with
            | nil => true
            | (k,v)::pm' =>
              match assoc mm k with
              | None => false
              | Some v' =>
                if psubmsg v bs v'
                then f pm'
                else false
              end
            end
        in f pm
      |  _, _ => false
      end.

    (* Submsg is reflexive. *)
    Lemma submsg_refl : forall x:msg, submsg x x = true.
    Admitted.

    (* Using bindings from a map match results in a submsg. *)
    Lemma patmatch_submsg_maps :
      forall l: list (string * pat),
      forall l0 : list (string * msg),
      forall bs : bindings,
        In bs (patmatch (PMap l) nil (Map l0)) ->
        psubmsg (PMap l) bs (Map l0) = true.
    Admitted.

    (* Big theorem #1: Bindings from a patmatch given a submsg. *)
    Theorem patmatch_submsg :
      forall p:pat,
      forall m:msg,
      forall bs:bindings,
        In bs (patmatch p nil m) ->
        psubmsg p bs m = true.
    Proof.
      intros.
      induction p.
      {
        induction m.
        {
          unfold psubmsg.
          destruct (string_dec s s0).
          {
            reflexivity.
          }
          {
            unfold patmatch in H.
            destruct (string_dec s s0).
            {
              unfold In in H.
              intuition.
            }
            {
              unfold In in H.
              contradiction.
            }
          }
        }
        {
          unfold psubmsg.
          unfold patmatch in H.
          unfold In in H.
          contradiction.
        }
      }
      {
        induction m.
        {
          unfold psubmsg.
          unfold patmatch in H.
          unfold assoc in H.
          unfold acons in H.
          unfold In in H.
          intuition.
          rewrite <- H0.
          simpl.
          destruct (string_dec s s).
          {
            simpl.
            destruct (string_dec s0 s0).
            reflexivity.
            intuition.
          }
          {
            intuition.
          }
        }
        {
          unfold psubmsg.
          unfold patmatch in H.
          unfold assoc in H.
          unfold acons in H.
          unfold In in H.
          intuition.
          rewrite <- H0.
          unfold assoc.
          destruct (string_dec s s).
          {
            apply submsg_refl.
          }
          {
            intuition.
          }
        }
      }
      {
        induction m.
        {
          unfold psubmsg.
          unfold patmatch in H.
          intuition.
        }
        {
          apply patmatch_submsg_maps in H.
          trivial.
        }
      }
    Qed.

    (* Big theorem #2: If some bindings give a submsg, then patmatch
    should find those bindings. *)
    Theorem submsg_patmatch :
      forall p:pat,
      forall m:msg,
      forall bs:bindings,
        psubmsg p bs m = true ->
        In bs (patmatch p nil m).
      Admitted.
    
  End verification.

End patmatch.


